package msgs

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
==================================================
		MarkUp
==================================================
*/

type MarkUp struct {
	Rows []Row
}

func NewMarkUp(rows ...Row) MarkUp {
	return MarkUp{
		Rows: rows,
	}
}

type Row struct {
	Buttons []Buttons
}

type Buttons interface {
	build(func(textKey string) string) tgbotapi.KeyboardButton
}

func NewRow(buttons ...Buttons) Row {
	return Row{
		Buttons: buttons,
	}
}

func (m MarkUp) Build(texts map[string]string) tgbotapi.ReplyKeyboardMarkup {
	var replyMarkUp tgbotapi.ReplyKeyboardMarkup

	for _, row := range m.Rows {
		replyMarkUp.Keyboard = append(replyMarkUp.Keyboard,
			row.buildRow(texts))
	}
	replyMarkUp.ResizeKeyboard = true
	return replyMarkUp
}

func (r Row) buildRow(texts map[string]string) []tgbotapi.KeyboardButton {
	var replyRow []tgbotapi.KeyboardButton

	for _, butt := range r.Buttons {
		replyRow = append(replyRow, butt.build(func(textKey string) string {
			return texts[textKey]
		}))
	}
	return replyRow
}

type DataButton struct {
	textKey string
}

func NewDataButton(key string) DataButton {
	return DataButton{
		textKey: key,
	}
}

func (b DataButton) build(langText func(textKey string) string) tgbotapi.KeyboardButton {
	text := langText(b.textKey)
	return tgbotapi.NewKeyboardButton(text)
}

type AdminButton struct {
	textKey string
}

func NewAdminButton(key string) AdminButton {
	return AdminButton{
		textKey: key,
	}
}

func (b AdminButton) build(adminText func(textKey string) string) tgbotapi.KeyboardButton {
	text := adminText(b.textKey)
	return tgbotapi.NewKeyboardButton(text)
}

/*
==================================================
		InlineMarkUp
==================================================
*/

type InlineMarkUp struct {
	Rows []InlineRow
}

func NewIlMarkUp(rows ...InlineRow) InlineMarkUp {
	return InlineMarkUp{
		Rows: rows,
	}
}

type InlineRow struct {
	Buttons []InlineButtons
}

type InlineButtons interface {
	build(func(textKey string) string) tgbotapi.InlineKeyboardButton
}

func NewIlRow(buttons ...InlineButtons) InlineRow {
	return InlineRow{
		Buttons: buttons,
	}
}

func (m InlineMarkUp) Build(texts map[string]string) tgbotapi.InlineKeyboardMarkup {
	var replyMarkUp tgbotapi.InlineKeyboardMarkup

	for _, row := range m.Rows {
		replyMarkUp.InlineKeyboard = append(replyMarkUp.InlineKeyboard,
			row.buildInlineRow(texts))
	}
	return replyMarkUp
}

func (r InlineRow) buildInlineRow(texts map[string]string) []tgbotapi.InlineKeyboardButton {
	var replyRow []tgbotapi.InlineKeyboardButton

	for _, butt := range r.Buttons {
		replyRow = append(replyRow, butt.build(func(textKey string) string {
			return texts[textKey]
		}))
	}
	return replyRow
}

type InlineDataButton struct {
	textKey string
	data    string
}

func NewIlDataButton(key, data string) InlineDataButton {
	return InlineDataButton{
		textKey: key,
		data:    data,
	}
}

func (b InlineDataButton) build(langText func(textKey string) string) tgbotapi.InlineKeyboardButton {
	text := langText(b.textKey)
	return tgbotapi.NewInlineKeyboardButtonData(text, b.data)
}

type InlineURLButton struct {
	textKey string
	url     string
}

func NewIlURLButton(key, url string) InlineURLButton {
	return InlineURLButton{
		textKey: key,
		url:     url,
	}
}

func (b InlineURLButton) build(langText func(textKey string) string) tgbotapi.InlineKeyboardButton {
	text := langText(b.textKey)
	return tgbotapi.NewInlineKeyboardButtonURL(text, b.url)
}

type InlineAdminButton struct {
	textKey string
	data    string
}

func NewIlAdminButton(key, data string) InlineAdminButton {
	return InlineAdminButton{
		textKey: key,
		data:    data,
	}
}

func (b InlineAdminButton) build(adminText func(textKey string) string) tgbotapi.InlineKeyboardButton {
	text := adminText(b.textKey)
	return tgbotapi.NewInlineKeyboardButtonData(text, b.data)
}

type InlineCustomButton struct {
	text string
	data string
}

func NewIlCustomButton(text, data string) InlineCustomButton {
	return InlineCustomButton{
		text: text,
		data: data,
	}
}

func (b InlineCustomButton) build(_ string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(b.text, b.data)
}
