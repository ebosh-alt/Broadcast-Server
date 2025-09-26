package http

import "time"

const (
	writeWait      = 10 * time.Second
	pongWait       = 30 * time.Second
	pingPeriod     = 20 * time.Second
	maxMessageSize = 4096 // 4 KiB
	sendBufSize    = 64
)
