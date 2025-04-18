package common

import (
	"time"
)

type Stopwatch struct {
	startTime time.Time
	running   bool
}

func (s *Stopwatch) Start() {
	s.startTime = time.Now()
	s.running = true
}

func (s *Stopwatch) Stop() time.Duration {
	if !s.running {
		return 0
	}
	s.running = false
	return time.Since(s.startTime)
}

func (s *Stopwatch) Reset() {
	s.startTime = time.Time{}
	s.running = false
}
