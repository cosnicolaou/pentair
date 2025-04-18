// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package slnet

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"time"

	"cloudeng.io/logging/ctxlog"
)

type Conn struct {
	conn    *net.TCPConn
	addr    string
	timeout time.Duration
}

type MessageHeader []byte

const MessageHeaderSize = 8

func (m MessageHeader) ID() uint16 {
	return binary.LittleEndian.Uint16(m[0:2])
}

func (m MessageHeader) Code() uint16 {
	return binary.LittleEndian.Uint16(m[2:4])
}

func (m MessageHeader) Size() uint32 {
	return binary.LittleEndian.Uint32(m[4:8])
}

func (m MessageHeader) Payload() []byte {
	return m[8:]
}

func (m MessageHeader) SetID(id uint16) {
	binary.LittleEndian.PutUint16(m[0:2], id)
}

func (m MessageHeader) SetCode(code uint16) {
	binary.LittleEndian.PutUint16(m[2:4], code)
}

func (m MessageHeader) SetSize(size uint32) {
	binary.LittleEndian.PutUint32(m[4:8], size)
}

func Dial(ctx context.Context, addr string, timeout time.Duration) (*Conn, error) {
	ctxlog.Info(ctx, "screenlogic: dialing", "addr", addr)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		ctxlog.Error(ctx, "screenlogic: dial failed", "addr", addr, "err", err)
		return nil, err
	}
	return &Conn{conn: conn.(*net.TCPConn), addr: addr, timeout: timeout}, nil
}

func (tc *Conn) Send(ctx context.Context, buf []byte) (int, error) {
	if err := tc.conn.SetWriteDeadline(time.Now().Add(tc.timeout)); err != nil {
		ctxlog.Error(ctx, "screenlogic: send failed to set read deadline", "addr", tc.addr, "err", err)
		return -1, err
	}
	n, err := tc.conn.Write(buf)
	hdr := MessageHeader(buf)
	ctxlog.Info(ctx, "screenlogic: sent", "addr", tc.addr, "id", hdr.ID(), "code", hdr.Code(), "size", hdr.Size(), "err", err)
	return n, err
}

func (tc *Conn) SendSensitive(ctx context.Context, buf []byte) (int, error) {
	return tc.Send(ctx, buf)
}

func (tc *Conn) readResponse() (MessageHeader, []byte, error) {
	buf := make([]byte, 1024)
	n, err := io.ReadAtLeast(tc.conn, buf, MessageHeaderSize)
	if err != nil {
		return nil, nil, err
	}
	hdr := MessageHeader(buf)
	msgSize := hdr.Size()
	if n < int(msgSize)+MessageHeaderSize {
		return nil, nil, io.ErrUnexpectedEOF
	}
	return hdr, buf[:n], nil
}

func (tc *Conn) ReadUntil(ctx context.Context, _ []string) ([]byte, error) {
	if err := tc.conn.SetReadDeadline(time.Now().Add(tc.timeout)); err != nil {
		ctxlog.Error(ctx, "screenlogic: readUntil failed to set read deadline", "addr", tc.addr, "err", err)
		return nil, err
	}
	hdr, buf, err := tc.readResponse()
	if err != nil {
		ctxlog.Error(ctx, "screenlogic: readUntil failed", "addr", tc.addr, "err", err)
		return nil, err
	}
	ctxlog.Info(ctx, "screenlogic: readUntil", "addr", tc.addr, "id", hdr.ID(), "code", hdr.Code(), "size", hdr.Size())
	return buf, err
}

func (tc *Conn) Close(ctx context.Context) error {
	if err := tc.conn.Close(); err != nil {
		ctxlog.Error(ctx, "screenlogic: close failed", "addr", tc.addr, "err", err)
		return err
	}
	ctxlog.Info(ctx, "screenlogic: close", "addr", tc.addr)
	return nil
}
