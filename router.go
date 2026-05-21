// router.go
package main

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// тупо выбор какую команду задействовать, для данного проекта можно было handlers.go не создавать, но при маштабировании, если оно будет-полезно, так что сделал в 2 файлах
func Router(ctx context.Context, bot *tgbotapi.BotAPI, cmd *Command) (string, error) {
	switch cmd.Name {
	// хеллоу
	case "start", "help":
		return HandleHelp(), nil
	// расход
	case "add":
		return HandleAdd(ctx, cmd.Args, cmd.UserID)
	// делит
	case "del":
		return HandleDelete(ctx, cmd.Args, cmd.UserID)
	// сумма рассходов
	case "calc":
		return HandleCalc(ctx, cmd.Args, cmd.UserID)
	// записи за период
	case "list":
		return HandleList(ctx, cmd.Args, cmd.UserID, "list")
	case "listweek", "listнеделя", "listн", "listw":
		return HandleList(ctx, cmd.Args, cmd.UserID, "week")
	case "listmonth", "listмесяц", "listм", "listm":
		return HandleList(ctx, cmd.Args, cmd.UserID, "month")
	case "listquarter", "listквартал", "listк", "listq":
		return HandleList(ctx, cmd.Args, cmd.UserID, "quarter")
	case "listhalfyear", "listполгода", "listп", "listh", "list6months", "list6 months", "list6месяцев", "list6 месяцев":
		return HandleList(ctx, cmd.Args, cmd.UserID, "halfyear")
	case "listyear", "listгод", "listг", "listy":
		return HandleList(ctx, cmd.Args, cmd.UserID, "year")
	case "listall", "listвсе", "lista", "listвсе время":
		return HandleList(ctx, cmd.Args, cmd.UserID, "all")
	// инфо
	case "info":
		return InfoAuthor(), nil
	//жалоба
	case "complaint":
		return HandleComplaint(ctx, bot, cmd.Args, cmd.UserID)

	//админ комнады
	case "admin":
		return HandleAdmin(ctx, cmd.Args, cmd.UserID)
	case "adminhelp":
		return HandleAdminHelp(), nil
	// если команды нет
	default:
		return "", fmt.Errorf("команда /%s не найдена. Напиши /help", cmd.Name)
	}
}
