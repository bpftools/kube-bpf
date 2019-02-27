/*
 * Copyright (C) 2017 Sylvain Afchain
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package insanelock

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var Activated bool
var Timeout = time.Second * 30

type RWMutex struct {
	mutex  sync.RWMutex
	frames atomic.Value
	at     atomic.Value
}

func (i *RWMutex) rwlock(l func()) {
	if !Activated {
		l()
		return
	}

	buffer := make([]byte, 10000)
	n := runtime.Stack(buffer, false)

	got := make(chan bool)
	go func() {
		select {
		case <-got:
		case <-time.After(Timeout):
			err := fmt.Sprintf("\n-- POTENTIAL DEADLOCK --\n\n")
			err += fmt.Sprintf("--  HOLDING THE LOCK SINCE %s  --\n", i.at.Load())
			err += fmt.Sprintf("%s\n", i.frames.Load())
			err += fmt.Sprintf("--  TRYING TO LOCK at %s  --\n", time.Now())
			err += fmt.Sprintf("%s\n", string(buffer[:n]))
			err += fmt.Sprintf("\n-- POTENTIAL DEADLOCK --\n")
			panic(err)
		}
	}()

	l()
	i.at.Store(time.Now())

	// stop the timer
	got <- true

	// save the current stack
	i.frames.Store(string(buffer[:n]))
}

func (i *RWMutex) Lock() {
	i.rwlock(i.mutex.Lock)
}

func (i *RWMutex) Unlock() {
	i.frames.Store("")
	i.mutex.Unlock()
}

func (i *RWMutex) RLock() {
	i.rwlock(i.mutex.RLock)
}

func (i *RWMutex) RUnlock() {
	i.mutex.RUnlock()
}
