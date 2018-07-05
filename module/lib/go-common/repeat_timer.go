// Copyright 2017 Annchain Information Technology Services Co.,Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import "time"
import "sync"

/*
RepeatTimer repeatedly sends a struct{}{} to .Ch after each "dur" period.
It's good for keeping connections alive.
A RepeatTimer must be Stop()'d or it will keep a goroutine alive.
*/
type RepeatTimer struct {
	Ch chan time.Time

	mtx    sync.Mutex
	name   string
	ticker *time.Ticker
	quit   chan struct{}
	done   chan struct{}
	dur    time.Duration
}

func NewRepeatTimer(name string, dur time.Duration) *RepeatTimer {
	var t = &RepeatTimer{
		Ch:     make(chan time.Time),
		ticker: time.NewTicker(dur),
		quit:   make(chan struct{}),
		done:   make(chan struct{}),
		name:   name,
		dur:    dur,
	}
	go t.fireRoutine(t.ticker)
	return t
}

func (t *RepeatTimer) fireRoutine(ticker *time.Ticker) {
	for {
		select {
		case t_ := <-ticker.C:
			t.Ch <- t_
		case <-t.quit:
			// needed so we know when we can reset t.quit
			t.done <- struct{}{}
			return
		}
	}
}

// Wait the duration again before firing.
func (t *RepeatTimer) Reset() {
	t.Stop()

	t.mtx.Lock() // Lock
	defer t.mtx.Unlock()

	t.ticker = time.NewTicker(t.dur)
	t.quit = make(chan struct{})
	go t.fireRoutine(t.ticker)
}

// For ease of .Stop()'ing services before .Start()'ing them,
// we ignore .Stop()'s on nil RepeatTimers.
func (t *RepeatTimer) Stop() bool {
	if t == nil {
		return false
	}
	t.mtx.Lock() // Lock
	defer t.mtx.Unlock()

	exists := t.ticker != nil
	if exists {
		t.ticker.Stop() // does not close the channel
		close(t.quit)
		<-t.done
		t.ticker = nil
	}
	return exists
}
