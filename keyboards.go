// keyboards.go
package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// возвращает основную инлайн-клавиатуру бота
func getMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить", "cmd_add"),
			tgbotapi.NewInlineKeyboardButtonData("📋 Список", "cmd_list"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Посчитать", "cmd_calc"),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить", "cmd_del"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❓ Помощь", "cmd_help"),
		),
	)
}

// возвращает клавиатуру выбора периода для команды /calc
// Используется только после нажатия кнопки «Посчитать»
func getCalcKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("День", "calc_day"),
			tgbotapi.NewInlineKeyboardButtonData("Неделя", "calc_week"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Месяц", "calc_month"),
			tgbotapi.NewInlineKeyboardButtonData("Квартал", "calc_quarter"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Полгода", "calc_halfyear"),
			tgbotapi.NewInlineKeyboardButtonData("Год", "calc_year"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Всё время", "calc_all"),
		),
	)
}

// Кнопки удаления для каждой записи
func createDeleteKeyboard(records []struct {
	ID   int
	Text string
}) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, r := range records {
		btn := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("🗑 %d", r.ID),
			fmt.Sprintf("del_%d", r.ID),
		)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	// Кнопки возврата
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Главное меню", "main_menu"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
