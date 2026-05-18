// router.go
package main

import (
	"context"
	"fmt"
)

// тупо выбор какую команду задействовать, для данного проекта можно было handlers.go не создавать, но при маштабировании, если оно будет-полезно, так что сделал в 2 файлах
func Router(ctx context.Context, cmd *Command) (string, error) {
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
	case "listweek":
		return HandleList(ctx, cmd.Args, cmd.UserID, "week")
	case "listmonth":
		return HandleList(ctx, cmd.Args, cmd.UserID, "month")
	case "listquarter":
		return HandleList(ctx, cmd.Args, cmd.UserID, "quarter")
	case "listhalfyear":
		return HandleList(ctx, cmd.Args, cmd.UserID, "halfyear")
	case "listyear":
		return HandleList(ctx, cmd.Args, cmd.UserID, "year")
	case "listall":
		return HandleList(ctx, cmd.Args, cmd.UserID, "all")
	// инфо
	case "info":
		return InfoAuthor(), nil
	// если команды нет
	default:
		return "", fmt.Errorf("команда /%s не найдена", cmd.Name)
	}
}
