// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"sync"

	"github.com/cosnicolaou/automation/net/netutil"
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

func NewSession(conn *slnet.Conn, idle netutil.IdleReset) Session {
	return &session{
		Session: streamconn.NewSession(conn, idle),
	}
}

func (s *session) NextID() uint16 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id++
	return s.id
}

type errorSession struct {
	streamconn.Session
}

func NewErrorSession(err error) Session {
	return &errorSession{
		Session: streamconn.NewErrorSession(err),
	}
}

func (e *errorSession) NextID() uint16 {
	return 0
}
