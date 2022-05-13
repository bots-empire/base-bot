package mailing

import (
	"fmt"
	"time"

	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	GlobalMailing = 4
)

type MailingUser struct {
	ID            int64
	Language      string
	AdvertChannel int
}

func (s *Service) StartMailing(botLang string, initiatorID int64, channel int) {
	startTime := time.Now()
	s.fillMessageMap()

	var (
		sendToUsers  int
		blockedUsers int
	)

	s.messages.SendNotificationToDeveloper(
		fmt.Sprintf("%s // mailing started", botLang),
		false,
	)

	for offset := 0; ; offset += s.usersPerIteration {
		countSend, errCount := s.mailToUserWithPagination(botLang, offset, channel)
		if countSend == -1 {
			s.sendRespMsgToMailingInitiator(initiatorID, "failing_mailing_text", sendToUsers)
			break
		}

		if countSend == 0 && errCount == 0 {
			break
		}

		sendToUsers += countSend
		blockedUsers += errCount
	}

	s.messages.SendNotificationToDeveloper(
		fmt.Sprintf("%s // send to %d users mail; latency: %v", botLang, sendToUsers, time.Now().Sub(startTime)),
		false,
	)

	s.sendRespMsgToMailingInitiator(initiatorID, "complete_mailing_text", sendToUsers)

	s.messages.Sender.UpdateBlockedUsers(channel)
}

func (s *Service) sendRespMsgToMailingInitiator(userID int64, key string, countOfSends int) {
	lang := s.messages.Sender.AdminLang(userID)
	text := fmt.Sprintf(s.messages.Sender.AdminText(lang, key), countOfSends)

	_ = s.messages.NewParseMessage(userID, text)
}

func (s *Service) mailToUserWithPagination(botLang string, offset int, channel int) (int, int) {
	users, err := s.getUsersWithPagination(offset)
	if err != nil {
		s.messages.SendNotificationToDeveloper(
			errors.Wrap(err, "get users with pagination").Error(),
			false,
		)

		return -1, 0
	}

	totalCount := len(users)
	if totalCount == 0 {
		return 0, 0
	}

	responseChan := make(chan bool)
	var sendToUsers int

	for _, user := range users {
		go s.sendMailToUser(botLang, user, responseChan, channel)
	}

	for countOfResp := 0; countOfResp < len(users); countOfResp++ {
		select {
		case resp := <-responseChan:
			if resp {
				sendToUsers++
			}
		}
	}

	return sendToUsers, totalCount - sendToUsers
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

func (s *Service) sendMailToUser(botLang string, user *MailingUser, respChan chan<- bool, channel int) {
	if channel == GlobalMailing {
		channel = user.AdvertChannel
	}

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", s.messages.Sender.GetAdvertURL(botLang, channel))),
	).Build(s.messages.Sender.GetTexts(user.Language))
	button := &markUp

	if !s.messages.Sender.ButtonUnderAdvert() {
		button = nil
	}

	baseChat := tgbotapi.BaseChat{
		ChatID:      user.ID,
		ReplyMarkup: button,
	}

	switch s.messages.Sender.AdvertisingChoice(channel) {
	case "photo":
		msg := s.photoMessageConfig[channel]
		//msg.BaseChat.ChatID = user.ID
		msg.BaseChat = baseChat
		respChan <- s.messages.SendMsgToUser(msg) == nil
	case "video":
		msg := s.videoMessageConfig[channel]
		//msg.BaseChat.ChatID = user.ID
		msg.BaseChat = baseChat
		respChan <- s.messages.SendMsgToUser(msg) == nil
	default:
		msg := s.messageConfigs[channel]
		//msg.BaseChat.ChatID = user.ID
		msg.BaseChat = baseChat
		respChan <- s.messages.SendMsgToUser(msg) == nil
	}
}

func (s *Service) fillMessageMap() {
	for _, lang := range s.messages.Sender.AvailableLang() {
		for i := 1; i < 6; i++ {
			//var markUp tgbotapi.InlineKeyboardMarkup
			text := s.messages.Sender.GetAdvertText(lang, i)

			s.nilConfig()

			//if !s.messages.Sender.ButtonUnderAdvert() {
			//	markUp = tgbotapi.InlineKeyboardMarkup{}
			//} else {
			//	markUp = msgs.NewIlMarkUp(
			//		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", s.messages.Sender.GetAdvertURL(lang, i))),
			//	).Build(s.messages.Sender.GetTexts(lang))
			//}

			switch s.messages.Sender.AdvertisingChoice(i) {
			case "photo":
				s.photoMessageConfig[i] = tgbotapi.PhotoConfig{
					BaseFile: tgbotapi.BaseFile{
						//BaseChat: tgbotapi.BaseChat{
						//	ReplyMarkup: markUp,
						//},
						File: tgbotapi.FileID(s.messages.Sender.GetAdvertisingPhoto(lang, i)),
					},
					Caption:   text,
					ParseMode: "HTML",
				}
			case "video":
				s.videoMessageConfig[i] = tgbotapi.VideoConfig{
					BaseFile: tgbotapi.BaseFile{
						//BaseChat: tgbotapi.BaseChat{
						//	ReplyMarkup: markUp,
						//},
						File: tgbotapi.FileID(s.messages.Sender.GetAdvertisingVideo(lang, i)),
					},
					Caption:   text,
					ParseMode: "HTML",
				}
			default:
				s.messageConfigs[i] = tgbotapi.MessageConfig{
					//BaseChat: tgbotapi.BaseChat{
					//	ReplyMarkup: markUp,
					//},
					Text: text,
				}
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
