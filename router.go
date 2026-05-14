// router.go
package main

import "fmt"

// тупо выбор какую команду задействовать, для данного проекта можно было handlers.go не создавать, но при маштабировании, если оно будет-полезно, так что сделал в 2 файлах
func Router(cmd *Command) (string, error) {
	switch cmd.Name {
	// хеллоу
	case "start", "help":
		return HandleHelp(), nil

	// расход
	case "add":
		return HandleAdd(cmd.Args, cmd.UserID)

	// делит
	case "del":
		return HandleDelete(cmd.Args, cmd.UserID)

	// сумма рассходов
	case "calc":
		return HandleCalc(cmd.Args, cmd.UserID)

	// записи за период
	case "list":
		return HandleList(cmd.Args, cmd.UserID, "list")
	case "listweek":
		return HandleList(cmd.Args, cmd.UserID, "week")
	case "listmonth":
		return HandleList(cmd.Args, cmd.UserID, "month")
	case "listquarter":
		return HandleList(cmd.Args, cmd.UserID, "quarter")
	case "listhalfyear":
		return HandleList(cmd.Args, cmd.UserID, "halfyear")
	case "listyear":
		return HandleList(cmd.Args, cmd.UserID, "year")
	case "listall":
		return HandleList(cmd.Args, cmd.UserID, "all")

	// инфо
	case "info":
		return InfoAuthor(), nil

	// если команды нет
	default:
		return "", fmt.Errorf("команда /%s не найдена", cmd.Name)
	}
}
