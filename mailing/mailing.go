package mailing

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bots-empire/base-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	statusActive      = "active"
	statusDeleted     = "deleted"
	statusNeedMailing = "mailing"
)

type MailingUser struct {
	ID            int64
	Language      string
	AdvertChannel int
}

func (s *Service) startSenderHandler() {
	for {
		users, err := s.getUsersWithMailing()
		if err != nil {
			s.errorHandler(err)
			continue
		}

		if len(users) == 0 {
			s.stopHandler()
			continue
		}

		for _, user := range users {
			go s.sendMailToUser(user)
		}
	}
}

func (s *Service) getUsersWithMailing() ([]*MailingUser, error) {
	rows, err := s.messages.Sender.GetDataBase().Query(`
SELECT id, lang, advert_channel
	FROM users
WHERE status = ?
ORDER BY id
	LIMIT ?;`,
		statusNeedMailing,
		s.usersPerIteration)
	if err != nil {
		return nil, errors.Wrap(err, "failed execute query")
	}

	return s.readUsersFromRows(rows)
}

func (s *Service) readUsersFromRows(rows *sql.Rows) ([]*MailingUser, error) {
	var users []*MailingUser

	for rows.Next() {
		user := &MailingUser{}

		if err := rows.Scan(
			&user.ID,
			&user.Language,
			&user.AdvertChannel); err != nil {
			return nil, errors.Wrap(err, "failed scan row")
		}

		if s.messages.Sender.CheckAdmin(user.ID) && !s.debugMode {
			continue
		}

		users = append(users, user)
	}

	return users, nil
}

func (s *Service) errorHandler(err error) {
	s.messages.SendNotificationToDeveloper(fmt.Sprintf("%s  //  error in mailing: %s", s.messages.Sender.GetBotLang(), err), false)
	time.Sleep(3 * time.Second)
}

func (s *Service) stopHandler() {
	<-s.startSignaller
	s.messages.SendNotificationToDeveloper(fmt.Sprintf("%s  //  mailing handler started", s.messages.Sender.GetBotLang()), false)
}

func (s *Service) StartMailing() error {
	s.fillMessageMap()

	s.messages.SendNotificationToDeveloper(
		fmt.Sprintf("%s // mailing started", s.messages.Sender.GetBotLang()),
		false,
	)

	return s.markMailingUsers()
}

func (s *Service) markMailingUsers() error {
	_, err := s.messages.Sender.GetDataBase().Exec(`
UPDATE users 
	SET status = ? 
WHERE status = ?;`,
		statusNeedMailing,
		statusActive)
	if err != nil {
		return errors.Wrap(err, "failed execute query")
	}

	return nil
}

func (s *Service) sendMailToUser(user *MailingUser) {

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text",
			s.messages.Sender.GetAdvertURL(s.messages.Sender.GetBotLang(), user.AdvertChannel)),
		),
	).Build(s.messages.Sender.GetTexts(user.Language))
	button := &markUp

	if !s.messages.Sender.ButtonUnderAdvert() {
		button = nil
	}

	baseChat := tgbotapi.BaseChat{
		ChatID:      user.ID,
		ReplyMarkup: button,
	}

	var err error
	switch s.messages.Sender.AdvertisingChoice(user.AdvertChannel) {
	case "photo":
		msg := s.photoMessageConfig[user.AdvertChannel]
		msg.BaseChat = baseChat
		err = s.messages.SendMsgToUser(msg, user.ID)
	case "video":
		msg := s.videoMessageConfig[user.AdvertChannel]
		msg.BaseChat = baseChat
		err = s.messages.SendMsgToUser(msg, user.ID)
	default:
		msg := s.messageConfigs[user.AdvertChannel]
		msg.BaseChat = baseChat
		err = s.messages.SendMsgToUser(msg, user.ID)
	}

	if err != nil {
		s.messages.SendNotificationToDeveloper(fmt.Sprintf("error in send mailing to user: %s", err), false)
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
