// stats.go
package main

import (
	"sync/atomic"
	"time"
)

var (
	startTime       = time.Now()
	totalMessages   int64
	totalErrors     int64
	totalComplaints int64
)

func incrementMessages() {
	atomic.AddInt64(&totalMessages, 1)
}

func incrementErrors() {
	atomic.AddInt64(&totalErrors, 1)
}

func incrementComplaints() {
	atomic.AddInt64(&totalComplaints, 1)
}

func getUptime() string {
	d := time.Since(startTime)
	return d.Truncate(time.Second).String()
}
