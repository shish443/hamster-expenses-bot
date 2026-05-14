// types.go
package main

type Command struct {
	Name   string
	Args   []string
	UserID int64
	ChatID int64
	MsgID  int
}
