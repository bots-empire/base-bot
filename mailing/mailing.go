package mailing

import (
	"fmt"
	"time"

	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	GlobalMailing  = 4
	ActiveStatus   = "active"
	InactiveStatus = "inactive"
	DeletedStatus  = "deleted"
)

type MailingUser struct {
	ID            int64
	Language      string
	AdvertChannel int
	Status        string
}

func (s *Service) StartMailing(botLang string, initiatorID int64, channel int) {
	startTime := time.Now()

	s.messages.SendNotificationToDeveloper(
		fmt.Sprintf("%s // mailing started", botLang),
		false,
	)

	if channel != GlobalMailing {
		err := s.UpdateStatusChannel(ActiveStatus, channel, InactiveStatus)
		if err != nil {
			s.messages.SendNotificationToDeveloper("failed to update status",
				false,
			)
		}
	} else {
		err := s.UpdateStatusAll(ActiveStatus, InactiveStatus)
		if err != nil {
			s.messages.SendNotificationToDeveloper("failed to update status",
				false,
			)
		}
	}

	countSend, _ := s.mailToUserWithPagination(initiatorID)

	s.messages.SendNotificationToDeveloper(
		fmt.Sprintf("%s // send to %d users mail; latency: %v", botLang, countSend, time.Now().Sub(startTime)),
		false,
	)

	s.sendRespMsgToMailingInitiator(initiatorID, "complete_mailing_text", countSend)

	s.messages.Sender.UpdateBlockedUsers(channel)
}

func (s *Service) sendRespMsgToMailingInitiator(userID int64, key string, countOfSends int) {
	lang := s.messages.Sender.AdminLang(userID)
	text := fmt.Sprintf(s.messages.Sender.AdminText(lang, key), countOfSends)

	_ = s.messages.NewParseMessage(userID, text)
}

func (s *Service) mailToUserWithPagination(initiatorID int64) (int, int) {
	var countSend int
	var totalCount int

	for offset := 0; ; offset += s.usersPerIteration {
		sendToUsers, countUsers := s.Mailing()

		if sendToUsers == -1 {
			s.sendRespMsgToMailingInitiator(initiatorID, "failing_mailing_text", sendToUsers)
			break
		}

		if sendToUsers == 0 && countUsers-sendToUsers == 0 {
			break
		}

		countSend += sendToUsers
		totalCount += countUsers
	}

	return countSend, totalCount - countSend
}

func (s *Service) Mailing() (int, int) {
	users, err := s.getUsersWithPagination(s.usersPerIteration)
	if err != nil {
		s.messages.SendNotificationToDeveloper(
			errors.Wrap(err, "get users with pagination").Error(),
			false,
		)
	}

	totalCount := len(users)
	if totalCount == 0 {
		return 0, 0
	}

	sendToUsers := s.SendingMail(users)

	return sendToUsers, totalCount
}

func (s *Service) SendingMail(users []*MailingUser) int {
	s.fillMessageMap()
	responseChan := make(chan bool)
	var sendToUsers int

	for _, user := range users {
		if user.Status == InactiveStatus {
			go s.sendMailToUser(user, responseChan)
		}
	}

	for countOfResp := 0; countOfResp < len(users); countOfResp++ {
		select {
		case resp := <-responseChan:
			if resp {
				sendToUsers++
			}
		}
	}

	return sendToUsers
}

func (s *Service) UpdateStatusChannel(currentStatus string, channel int, status string) error {
	_, err := s.messages.Sender.GetDataBase().Exec(`
UPDATE users SET status = ? WHERE advert_channel = ? AND status = ?;`, status, channel, currentStatus)
	if err != nil {
		return errors.Wrap(err, "failed execute query")
	}

	return nil
}

func (s *Service) UpdateStatusChannelFromID(id int64, status string, channel int) error {
	_, err := s.messages.Sender.GetDataBase().Exec(`
UPDATE users SET status = ? WHERE advert_channel = ? AND id = ?;`, status, channel, id)
	if err != nil {
		return errors.Wrap(err, "failed execute query")
	}

	return nil
}

func (s *Service) UpdateStatusAll(currentStatus string, status string) error {
	_, err := s.messages.Sender.GetDataBase().Exec(`
UPDATE users SET status = ? WHERE status = ?;`, status, currentStatus)
	if err != nil {
		return errors.Wrap(err, "failed execute query")
	}

	return nil
}

func (s *Service) getUsersWithPagination(offset int) ([]*MailingUser, error) {
	rows, err := s.messages.Sender.GetDataBase().Query(`
SELECT id, lang, advert_channel 
	FROM users 
ORDER BY id 
	LIMIT ? 
	OFFSET ?;`,
		s.usersPerIteration,
		offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed execute query")
	}

	var users []*MailingUser

	for rows.Next() {
		user := &MailingUser{}

		if err := rows.Scan(&user.ID, &user.Language, &user.AdvertChannel); err != nil {
			return nil, errors.Wrap(err, "failed scan row")
		}

		if s.messages.Sender.CheckAdmin(user.ID) {
			continue
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *Service) sendMailToUser(user *MailingUser, respChan chan<- bool) {
	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", s.messages.Sender.GetAdvertURL(user.Language, user.AdvertChannel))),
	).Build(s.messages.Sender.GetTexts(user.Language))
	button := &markUp

	if !s.messages.Sender.ButtonUnderAdvert() {
		button = nil
	}

	baseChat := tgbotapi.BaseChat{
		ChatID:      user.ID,
		ReplyMarkup: button,
	}

	switch s.messages.Sender.AdvertisingChoice(user.AdvertChannel) {
	case "photo":
		msg := s.photoMessageConfig[user.AdvertChannel]
		msg.BaseChat = baseChat
		err := s.messages.SendMsgToUser(msg)
		if err != nil {
			err := s.messages.SendMsgToUser(msg)
			if err != nil {
				s.UpdateStatusChannelFromID(user.ID, DeletedStatus, user.AdvertChannel)
				respChan <- false
			}
		} else {
			s.UpdateStatusChannel(InactiveStatus, user.AdvertChannel, ActiveStatus)
			respChan <- true
		}
	case "video":
		msg := s.videoMessageConfig[user.AdvertChannel]
		msg.BaseChat = baseChat
		err := s.messages.SendMsgToUser(msg)
		if err != nil {
			err := s.messages.SendMsgToUser(msg)
			if err != nil {
				s.UpdateStatusChannelFromID(user.ID, DeletedStatus, user.AdvertChannel)
				respChan <- false
			}
		} else {
			s.UpdateStatusChannel(InactiveStatus, user.AdvertChannel, ActiveStatus)
			respChan <- true
		}
	default:
		msg := s.messageConfigs[user.AdvertChannel]
		msg.BaseChat = baseChat
		err := s.messages.SendMsgToUser(msg)
		if err != nil {
			err := s.messages.SendMsgToUser(msg)
			if err != nil {
				s.UpdateStatusChannelFromID(user.ID, DeletedStatus, user.AdvertChannel)
				respChan <- false
			}
		} else {
			s.UpdateStatusChannel(InactiveStatus, user.AdvertChannel, ActiveStatus)
			respChan <- true
		}
	}
}

func (s *Service) fillMessageMap() {
	lang := s.messages.Sender.GetBotLang()
	for i := 1; i < 6; i++ {
		text := s.messages.Sender.GetAdvertText(lang, i)

		s.nilConfig()

		switch s.messages.Sender.AdvertisingChoice(i) {
		case "photo":
			s.photoMessageConfig[i] = tgbotapi.PhotoConfig{
				BaseFile: tgbotapi.BaseFile{
					File: tgbotapi.FileID(s.messages.Sender.GetAdvertisingPhoto(lang, i)),
				},
				Caption:   text,
				ParseMode: "HTML",
			}
		case "video":
			s.videoMessageConfig[i] = tgbotapi.VideoConfig{
				BaseFile: tgbotapi.BaseFile{
					File: tgbotapi.FileID(s.messages.Sender.GetAdvertisingVideo(lang, i)),
				},
				Caption:   text,
				ParseMode: "HTML",
			}
		default:
			s.messageConfigs[i] = tgbotapi.MessageConfig{
				Text: text,
			}
		}
	}
}

func (s *Service) nilConfig() {
	if s.messageConfigs == nil || s.photoMessageConfig == nil || s.videoMessageConfig == nil {
		s.messageConfigs = make(map[int]tgbotapi.MessageConfig, 10)
		s.photoMessageConfig = make(map[int]tgbotapi.PhotoConfig, 10)
		s.videoMessageConfig = make(map[int]tgbotapi.VideoConfig, 10)
	}
}
