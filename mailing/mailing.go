package mailing

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"

	"github.com/bots-empire/base-bot/msgs"
)

const (
	statusActive      = "active"
	statusNeedMailing = "mailing"
	statusInitMailing = "init_mailing"
)

type MailingUser struct {
	ID            int64
	Language      string
	AdvertChannel int
}

func (s *Service) startSenderHandler() {
	s.fillMessageMap()

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

		wg := &sync.WaitGroup{}
		wg.Add(len(users))

		for _, user := range users {
			go s.sendMailToUser(wg, user)
		}

		wg.Wait()
	}
}

func (s *Service) getUsersWithMailing() ([]*MailingUser, error) {
	rows, err := s.messages.Sender.GetDataBase().Query(
		renderSQL("get_users", s.messages.Sender.GetRelationName(), s.dbType),
		statusNeedMailing,
		s.usersPerIteration)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed execute query in get users with pagination, per inter = %d", s.usersPerIteration))
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

func (s *Service) sendErrorToAdmin(err error) {
	s.messages.SendNotificationToDeveloper(fmt.Sprintf("%s  //  error in mailing: %s", s.messages.Sender.GetBotLang(), err), false)
}

func (s *Service) stopHandler() {
	userIDs, err := s.getUsersWithInitMailing()
	if err != nil {
		s.sendErrorToAdmin(err)
	}

	err, count := s.countMailingUsers()
	if err != nil {
		s.messages.SendNotificationToDeveloper(fmt.Sprintf("failed count mailing users: %s", err), false)
	}

	for _, user := range userIDs {
		err = s.messages.NewParseMessage(user.ID, fmt.Sprintf("%s // mailing completed // Total : %d", s.messages.Sender.GetBotLang(), count))
		s.messages.SendNotificationToDeveloper(fmt.Sprintf("err in new parse message: %s", err), false)

		err = s.markReadyMailingUser(user.ID)
		s.sendErrorToAdmin(err)
	}

	<-s.startSignaller
	if s.debugMode {
		s.messages.SendNotificationToDeveloper(fmt.Sprintf("%s  //  mailing handler started", s.messages.Sender.GetBotLang()), false)
	}
}

func (s *Service) countMailingUsers() (error, int) {
	rows, err := s.messages.Sender.GetDataBase().Query(
		renderSQL("count_mailing_users", s.messages.Sender.GetRelationName(), s.dbType),
		statusActive)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed execute query in get users with pagination, per inter = %d", s.usersPerIteration)), 0
	}

	count, err := s.readCountMailingUsersRows(rows)
	if err != nil {
		return errors.Wrap(err, "failed read Users From Rows"), 0
	}

	return nil, count
}

func (s *Service) readCountMailingUsersRows(rows *sql.Rows) (int, error) {
	defer rows.Close()

	var count int

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, errors.Wrap(err, "failed to scan row")
		}
	}

	return count, nil
}

func (s *Service) getUsersWithInitMailing() ([]*MailingUser, error) {
	rows, err := s.messages.Sender.GetDataBase().Query(
		renderSQL("get_users", s.messages.Sender.GetRelationName(), s.dbType),
		statusInitMailing,
		s.usersPerIteration)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed execute query in get users with pagination, per inter = %d", s.usersPerIteration))
	}

	return s.readUsersFromRows(rows)
}

func (s *Service) StartMailing(channels []int, id int64) error {
	s.fillMessageMap()
	err := s.markInitMailingUsers(id)
	if err != nil {
		return err
	}

	if s.debugMode {
		s.messages.SendNotificationToDeveloper(
			fmt.Sprintf("%s // mailing started", s.messages.Sender.GetBotLang()),
			false,
		)
	}

	for _, userChannel := range channels {
		err = s.markMailingUsers(userChannel)
		if err != nil {
			return err
		}
	}

	s.startSignaller <- true

	return nil
}

func (s *Service) markMailingUsers(usersChan int) error {
	_, err := s.messages.Sender.GetDataBase().Exec(
		renderSQL("mark_mailing_user", s.messages.Sender.GetRelationName(), s.dbType),
		statusNeedMailing,
		statusActive,
		usersChan)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed execute query in mark mailing users, users chan = %d", usersChan))
	}

	return nil
}

func (s *Service) markInitMailingUsers(id int64) error {
	_, err := s.messages.Sender.GetDataBase().Exec(
		renderSQL("mark_init_mailing_user", s.messages.Sender.GetRelationName(), s.dbType),
		statusInitMailing,
		id)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed execute query in mark init mailing users, users chan = %d", id))
	}

	return nil
}

func (s *Service) sendMailToUser(wg *sync.WaitGroup, user *MailingUser) {
	defer wg.Done()

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

	var (
		err  error
		send bool
	)

	switch s.messages.Sender.AdvertisingChoice(user.AdvertChannel) {
	case "photo":
		msg := s.photoMessageConfig[user.AdvertChannel]
		msg.BaseChat = baseChat
		err, send = s.messages.SendMailToUser(msg, user.ID)
	case "video":
		msg := s.videoMessageConfig[user.AdvertChannel]
		msg.BaseChat = baseChat
		err, send = s.messages.SendMailToUser(msg, user.ID)
	default:
		msg := s.messageConfigs[user.AdvertChannel]
		msg.BaseChat = baseChat
		err, send = s.messages.SendMailToUser(msg, user.ID)
	}

	if !send {
		s.messages.Sender.GetMetrics("total_block_users").WithLabelValues(s.messages.Sender.GetBotLang()).Inc()
		return
	}

	if err != nil {
		s.sendErrorToAdmin(err)
		return
	}

	if err = s.markReadyMailingUser(user.ID); err != nil {
		s.sendErrorToAdmin(err)
	}

	s.messages.Sender.GetMetrics("total_mailing_users").WithLabelValues(s.messages.Sender.GetBotLang()).Inc()
}

func (s *Service) markReadyMailingUser(userID int64) error {
	_, err := s.messages.Sender.GetDataBase().Exec(
		renderSQL("mark_active_user", s.messages.Sender.GetRelationName(), s.dbType),
		statusActive,
		userID)
	if err != nil {
		return errors.Wrap(err, "failed execute query in mark ready user")
	}

	return nil
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
