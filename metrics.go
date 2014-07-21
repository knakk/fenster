package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rcrowley/go-metrics"
)

type appMetrics struct {
	StartTime time.Time
	PID       int
}

type exportMetrics struct {
	UpTime  string
	PID     int
	Metrics metrics.Registry
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
		UpTime:  uptime.String(),
		PID:     m.PID,
		Metrics: metrics.DefaultRegistry,
	}
}

// The following code is from copied from
// https://github.com/rcrowley/go-tigertonic/blob/master/metrics.go

// CounterByStatusXX is an http.Handler that counts responses by the first
// digit of their HTTP status code via go-metrics.
type CounterByStatusXX struct {
	counter1xx, counter2xx, counter3xx, counter4xx, counter5xx metrics.Counter
	handler                                                    http.Handler
}

// CountedByStatusXX returns an http.Handler that passes requests to an
// underlying http.Handler and then counts the response by the first digit of
// its HTTP status code via go-metrics.
func CountedByStatusXX(
	handler http.Handler,
	name string,
	registry metrics.Registry,
) *CounterByStatusXX {
	if nil == registry {
		registry = metrics.DefaultRegistry
	}
	c := &CounterByStatusXX{
		counter1xx: metrics.NewCounter(),
		counter2xx: metrics.NewCounter(),
		counter3xx: metrics.NewCounter(),
		counter4xx: metrics.NewCounter(),
		counter5xx: metrics.NewCounter(),
		handler:    handler,
	}
	if err := registry.Register(
		fmt.Sprintf("%s-1xx", name),
		c.counter1xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-2xx", name),
		c.counter2xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-3xx", name),
		c.counter3xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-4xx", name),
		c.counter4xx,
	); nil != err {
		panic(err)
	}
	if err := registry.Register(
		fmt.Sprintf("%s-5xx", name),
		c.counter5xx,
	); nil != err {
		panic(err)
	}
	return c
}

// ServeHTTP passes the request to the underlying http.Handler and then counts
// the response by its HTTP status code via go-metrics.
func (c *CounterByStatusXX) ServeHTTP(w0 http.ResponseWriter, r *http.Request) {
	w := NewTeeHeaderResponseWriter(w0)
	c.handler.ServeHTTP(w, r)
	if w.StatusCode < 200 {
		c.counter1xx.Inc(1)
	} else if w.StatusCode < 300 {
		c.counter2xx.Inc(1)
	} else if w.StatusCode < 400 {
		c.counter3xx.Inc(1)
	} else if w.StatusCode < 500 {
		c.counter4xx.Inc(1)
	} else {
		c.counter5xx.Inc(1)
	}
}

// Timer is an http.Handler that counts requests via go-metrics.
type Timer struct {
	metrics.Timer
	handler http.Handler
}

// Timed returns an http.Handler that starts a timer, passes requests to an
// underlying http.Handler, stops the timer, and updates the timer via
// go-metrics.
func Timed(handler http.Handler, name string, registry metrics.Registry) *Timer {
	timer := &Timer{
		Timer:   metrics.NewTimer(),
		handler: handler,
	}
	if nil == registry {
		registry = metrics.DefaultRegistry
	}
	if err := registry.Register(name, timer); nil != err {
		panic(err)
	}
	return timer
}

// ServeHTTP starts a timer, passes the request to the underlying http.Handler,
// stops the timer, and updates the timer via go-metrics.
func (t *Timer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer t.UpdateSince(time.Now())
	t.handler.ServeHTTP(w, r)
}
