// stats.go
package main

import (
	"sync/atomic"
	"time"
)

// глобальные счётчики для мониторинга работы бота.
var (
	startTime       = time.Now() //момент запуска бота
	totalMessages   int64        //количество сообщений
	totalErrors     int64        //количество ошибок
	totalComplaints int64        //количество отправленых жалоб
)

// счетчик сообщений
func incrementMessages() {
	atomic.AddInt64(&totalMessages, 1)
}

// счетчик ошибок
func incrementErrors() {
	atomic.AddInt64(&totalErrors, 1)
}

// счетчик жалоб
func incrementComplaints() {
	atomic.AddInt64(&totalComplaints, 1)
}

// время работы в норм виде
func getUptime() string {
	d := time.Since(startTime)
	return d.Truncate(time.Second).String()
}
