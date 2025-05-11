// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"sync/atomic"

	"github.com/cosnicolaou/automation/net/streamconn"
)

type Session struct {
	*streamconn.Session
}

var id uint32

func NewSession(s *streamconn.Session) *Session {
	return &Session{
		Session: s,
	}
}

func (s *Session) NextID() uint16 {
	ui := atomic.AddUint32(&id, 1)
	return uint16(ui & 0xFFFF)
}
