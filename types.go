// types.go
package main

// тупо тип, думал их будет больше поэтому выделили файл
type Command struct {
	Name   string
	Args   []string
	UserID int64
	ChatID int64
	MsgID  int
}
