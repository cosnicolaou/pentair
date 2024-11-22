// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"fmt"
	"time"
)

var (
	connectMsg  = []byte("CONNECTSERVERHOST\r\n\r\n")
	loginClient = "automation"
	loginPasswd = "0000000000000000" // <= 16 bytes

)

func Login(ctx context.Context, s Session) error {
	id := s.NextID()

	// Build the login message which consists of:
	// int, int, client, password, int for which none of the
	// values seem to matter.
	size := StringSize(loginClient) + StringSize(loginPasswd) + 12
	loginMsg := NewEmptyMessage(id, MsgLocalLogin, size)
	pl := loginMsg.Payload()
	pl = AppendUint32(pl, 0)
	pl = AppendUint32(pl, 0)
	pl = AppendString(pl, loginClient)
	pl = AppendString(pl, loginPasswd)
	pl = AppendUint32(pl, 0)

	// Send the raw connect string to kick start the session
	// and then the login message.
	s.Send(ctx, connectMsg)
	s.Send(ctx, loginMsg)

	rm := Message(s.ReadUntil(ctx))
	if rm.Code() == MsgBadLogin {
		return fmt.Errorf("Connect: failed: bad login: %w", ErrBadLogin)
	}
	if err := ValidateResponse(s, rm, id, MsgLocalLogin); err != nil {
		return fmt.Errorf("Connect: failed: %w", err)
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("Connect: unexpected error: %w", err)
	}
	return nil
}

func GetTimeAndDate(ctx context.Context, s Session) (time.Time, error) {
	id := s.NextID()
	m := NewEmptyMessage(id, MsgGetDateTime, 0)
	rm, err := sendAndValidate(ctx, s, m, id, MsgGetDateTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("GetTimeAndDate: %w", err)
	}
	return DecodeDateTime(rm)
}

func GetVersionInfo(ctx context.Context, s Session) (string, error) {
	id := s.NextID()
	m := NewEmptyMessage(id, MsgGetVersion, 0)
	rm, err := sendAndValidate(ctx, s, m, id, MsgGetVersion)
	if err != nil {
		return "", fmt.Errorf("GetVersionInfo: %w", err)
	}
	return DecodeVersion(rm), nil
}
