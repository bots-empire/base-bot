package msgs

import (
	"strings"

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

	return s.SendMsgToUser(msg)
}

func (s *Service) NewIDParseMessage(chatID int64, text string) (int, error) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      s.insertCurrency(text),
		ParseMode: "HTML",
	}

	message, err := s.Sender.GetBot().Send(msg)
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

	return s.SendMsgToUser(msg)
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

	return s.SendMsgToUser(msg)
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

	return s.SendMsgToUser(msg)
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

	message, err := s.Sender.GetBot().Send(msg)
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

	return s.SendMsgToUser(msg)
}

func (s *Service) SendAnswerCallback(callbackQuery *tgbotapi.CallbackQuery, lang, text string) error {
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            s.Sender.LangText(lang, text),
	}

	_ = s.SendMsgToUser(answerCallback)
	return nil
}

func (s *Service) SendAdminAnswerCallback(callbackQuery *tgbotapi.CallbackQuery, text string) error {
	lang := s.Sender.AdminLang(callbackQuery.From.ID)
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            s.Sender.AdminText(lang, text),
	}

	_ = s.SendMsgToUser(answerCallback)
	return nil
}

func (s *Service) SendSimpleMsg(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, s.insertCurrency(text))

	return s.SendMsgToUser(msg)
}

func (s *Service) SendMsgToUser(msg tgbotapi.Chattable) error {
	if _, err := s.Sender.GetBot().Send(msg); err != nil {
		return err
	}
	return nil
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
	})
}

func (s *Service) insertCurrency(text string) string {
	return strings.Replace(text, currency, s.Sender.GetCurrency(), -1)
}
