package main

import (
	"os"
	"time"
)

type appMetrics struct {
	StartTime time.Time
	PID       int
}

type exportMetrics struct {
	UpTime string
	PID    int
}

func registerMetrics() *appMetrics {
	var m appMetrics

	m.StartTime = time.Now()
	m.PID = os.Getpid()

	return &m
}

func (m *appMetrics) Export() *exportMetrics {
	now := time.Now()
	uptime := now.Sub(m.StartTime)

	return &exportMetrics{
		UpTime: uptime.String(),
		PID:    m.PID,
	}
}
