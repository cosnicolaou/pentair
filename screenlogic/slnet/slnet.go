// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package slnet

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"time"
)

type Conn struct {
	conn    *net.TCPConn
	timeout time.Duration
	logger  *slog.Logger
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

func Dial(ctx context.Context, addr string, timeout time.Duration, logger *slog.Logger) (*Conn, error) {
	logger.Log(ctx, slog.LevelInfo, "dialing screenlogic", "addr", addr)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		logger.Log(ctx, slog.LevelWarn, "dial failed", "addr", addr, "err", err)
		return nil, err
	}
	logger = logger.With("protocol", "screenlogic", "addr", conn.RemoteAddr().String())
	c := &Conn{
		conn:    conn.(*net.TCPConn),
		timeout: timeout,
		logger:  logger}
	return c, nil
}

func (tc *Conn) Send(ctx context.Context, buf []byte) (int, error) {
	if err := tc.conn.SetWriteDeadline(time.Now().Add(tc.timeout)); err != nil {
		tc.logger.Log(ctx, slog.LevelWarn, "send failed to set read deadline", "err", err)
		return -1, err
	}
	n, err := tc.conn.Write(buf)
	hdr := MessageHeader(buf)
	tc.logger.Log(ctx, slog.LevelInfo, "sent", "id", hdr.ID(), "code", hdr.Code(), "size", hdr.Size(), "err", err)
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
		tc.logger.Log(ctx, slog.LevelWarn, "readUntil failed to set read deadline", "err", err)
		return nil, err
	}
	hdr, buf, err := tc.readResponse()
	if err != nil {
		tc.logger.Log(ctx, slog.LevelWarn, "readUntil failed", "err", err)
		return nil, err
	}
	tc.logger.Log(ctx, slog.LevelInfo, "readUntil", "id", hdr.ID(), "code", hdr.Code(), "size", hdr.Size())
	return buf, err
}

func (tc *Conn) Close(ctx context.Context) error {
	if err := tc.conn.Close(); err != nil {
		tc.logger.Log(ctx, slog.LevelWarn, "close failed", "err", err)
	}
	tc.logger.Log(ctx, slog.LevelInfo, "close")
	return nil
}
