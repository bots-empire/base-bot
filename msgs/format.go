package msgs

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	currency = "{{currency}}"
)

func (s *Service) NewParseMessage(chatID int64, text string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      s.insertCurrency(text),
		ParseMode: "HTML",
	}

	return s.SendMsgToUser(msg, chatID)
}

func (s *Service) NewIDParseMessage(chatID int64, text string) (int, error) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      s.insertCurrency(text),
		ParseMode: "HTML",
	}

	message, err := s.sendMsgToUser(msg, chatID)
	if err != nil {
		return 0, nil
	}
	return message.MessageID, nil
}

func (s *Service) NewParseMarkUpMessage(chatID int64, markUp interface{}, text string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      s.insertCurrency(text),
		ParseMode: "HTML",
	}

	return s.SendMsgToUser(msg, chatID)
}

func (s *Service) NewParseMarkUpPhotoMessage(chatID int64, markUp interface{}, text string, photo tgbotapi.RequestFileData) error {
	msg := tgbotapi.PhotoConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatID:      chatID,
				ReplyMarkup: markUp,
			},
			File: photo},
		Caption:   s.insertCurrency(text),
		ParseMode: "HTML",
	}

	return s.SendMsgToUser(msg, chatID)
}

func (s *Service) NewParseMarkUpVideoMessage(chatID int64, markUp interface{}, text string, video tgbotapi.RequestFileData) error {
	msg := tgbotapi.VideoConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatID:      chatID,
				ReplyMarkup: markUp,
			},
			File: video,
		},
		Caption:   s.insertCurrency(text),
		ParseMode: "HTML",
	}

	return s.SendMsgToUser(msg, chatID)
}

func (s *Service) NewIDParseMarkUpMessage(chatID int64, markUp interface{}, text string) (int, error) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:                  s.insertCurrency(text),
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}

	message, err := s.sendMsgToUser(msg, chatID)
	if err != nil {
		return 0, err
	}
	return message.MessageID, nil
}

func (s *Service) NewEditMarkUpMessage(userID int64, msgID int, markUp *tgbotapi.InlineKeyboardMarkup, text string) error {
	msg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      userID,
			MessageID:   msgID,
			ReplyMarkup: markUp,
		},
		Text:                  s.insertCurrency(text),
		ParseMode:             "HTML",
		DisableWebPagePreview: true,
	}

	return s.SendMsgToUser(msg, userID)
}

func (s *Service) SendAnswerCallback(callbackQuery *tgbotapi.CallbackQuery, text string) error {
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            text,
	}

	_ = s.SendMsgToUser(answerCallback, callbackQuery.From.ID)
	return nil
}

func (s *Service) SendAdminAnswerCallback(callbackQuery *tgbotapi.CallbackQuery, text string) error {
	lang := s.Sender.AdminLang(callbackQuery.From.ID)
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            s.Sender.AdminText(lang, text),
	}

	_ = s.SendMsgToUser(answerCallback, callbackQuery.From.ID)
	return nil
}

func (s *Service) SendSimpleMsg(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, s.insertCurrency(text))

	return s.SendMsgToUser(msg, chatID)
}

func (s *Service) SendMsgToUser(msg tgbotapi.Chattable, userID int64) error {
	_, err := s.sendMsgToUser(msg, userID)
	return err
}

func (s *Service) sendMsgToUser(msg tgbotapi.Chattable, userID int64) (tgbotapi.Message, error) {
	var returnErr error

	for i := 0; i < 10; i++ {
		sendMsg, err := s.Sender.GetBot().Send(msg)
		if err == nil {
			return sendMsg, nil
		}

		if s.errorHandler(err, userID) {
			return tgbotapi.Message{}, nil
		}

		returnErr = err

		time.Sleep(time.Second)
	}

	return tgbotapi.Message{}, returnErr
}

func (s *Service) errorHandler(err error, userID int64) bool {
	fmt.Println(err)
	if err.Error() == "Forbidden: bot was blocked by the user" ||
		err.Error() == "Forbidden: bot can't initiate conversation with a user" ||
		err.Error() == "json: cannot unmarshal bool into Go value of type tgbotapi.Message" {
		if blockErr := s.Sender.BlockUser(userID); blockErr != nil {
			s.SendNotificationToDeveloper(blockErr.Error(), false)
		}

		return true
	}

	if strings.Contains(err.Error(), "Too Many Requests: retry after") {
		splitTime := time.Duration(getSleepTimeFromErr(err.Error()))
		time.Sleep(splitTime * time.Second)
	}

	return false
}

func getSleepTimeFromErr(err string) int {
	splitErr := strings.Split(err, " ")
	for i, word := range splitErr {
		if word == "after" {
			if len(splitErr) > i {
				num, _ := strconv.Atoi(splitErr[i+1])
				return num
			}
		}
	}

	return 0
}

func (s *Service) SendNotificationToDeveloper(text string, needPin bool) {
	for _, developerID := range s.Developers {
		msgID, _ := s.NewIDParseMessage(developerID, text)

		if needPin {
			s.PinMsgToDeveloper(developerID, msgID)
		}
	}
}

func (s *Service) PinMsgToDeveloper(userID int64, msgID int) {
	_ = s.SendMsgToUser(tgbotapi.PinChatMessageConfig{
		ChatID:    userID,
		MessageID: msgID,
	}, userID)
}

func (s *Service) insertCurrency(text string) string {
	return strings.Replace(text, currency, s.Sender.GetCurrency(), -1)
}
