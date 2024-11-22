// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"sync"

	"github.com/cosnicolaou/automation/net/streamconn"
	"github.com/cosnicolaou/pentair/screenlogic/slnet"
)

type Session interface {
	streamconn.Session
	NextID() uint16
}

type session struct {
	streamconn.Session
	mu sync.Mutex
	id uint16
}

func NewSession(conn *slnet.Conn, idle *streamconn.IdleTimer) Session {
	sess := streamconn.NewSession(conn, idle)
	return &session{
		Session: sess,
	}
}

func (s *session) NextID() uint16 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id++
	return s.id
}
