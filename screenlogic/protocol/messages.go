// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf16"

	"github.com/cosnicolaou/pentair/screenlogic/slnet"
)

type Message []byte

func (m Message) ID() uint16 {
	return slnet.MessageHeader(m).ID()
}

func (m Message) Code() MsgCode {
	return MsgCode(slnet.MessageHeader(m).Code())
}

func (m Message) Size() uint32 {
	return slnet.MessageHeader(m).Size()
}

func (m Message) Payload() []byte {
	return slnet.MessageHeader(m).Payload()
}

func (m Message) SetID(id uint16) {
	slnet.MessageHeader(m).SetID(id)
}

func (m Message) SetCode(code MsgCode) {
	slnet.MessageHeader(m).SetCode(uint16(code))
}

func (m Message) SetSize(size uint32) {
	slnet.MessageHeader(m).SetSize(size)
}

func NewMessage(id uint16, code MsgCode, payload []byte) Message {
	msg := Message(make([]byte, slnet.MessageHeaderSize+len(payload)))
	msg.SetID(id)
	msg.SetCode(code)
	msg.SetSize(uint32(len(payload)))
	copy(msg.Payload(), payload)
	return msg
}

func NewEmptyMessage(id uint16, code MsgCode, size uint32) Message {
	msg := Message(make([]byte, slnet.MessageHeaderSize+size))
	msg.SetID(id)
	msg.SetCode(code)
	msg.SetSize(size)
	return msg
}

func roundTo4(size int) uint32 {
	n := uint32(size)
	r := n % 4
	if r != 0 {
		r = 4 - r
	}
	return n + r
}

func BytesSize(msg []byte) uint32 {
	return roundTo4(len(msg)) + 4
}

func StringSize(s string) uint32 {
	return roundTo4(len(s)) + 4
}

// AppendBytes appends the message to the buffer and returns the remaining buffer.
// It rounds to the size of the message to the nearest multiple of 4 and
// prepends the size as a uint32.
func AppendBytes(buf, msg []byte) []byte {
	size := roundTo4(len(msg))
	binary.LittleEndian.PutUint32(buf, size)
	copy(buf[4:], msg)
	return buf[size+4:]
}

// AppendString appends the message to the buffer and returns the remaining buffer.
// It rounds to the size of the message to the nearest multiple of 4 and
// prepends the size as a uint32.
func AppendString(buf []byte, msg string) []byte {
	size := roundTo4(len(msg))
	binary.LittleEndian.PutUint32(buf, size)
	copy(buf[4:], msg)
	return buf[size+4:]
}

func AppendUint32(buf []byte, val uint32) []byte {
	binary.LittleEndian.PutUint32(buf, val)
	return buf[4:]
}

func AppendUint16(buf []byte, val uint16) []byte {
	binary.LittleEndian.PutUint16(buf, val)
	return buf[2:]
}

func DecodeUint8(buf []byte, ok bool, val *uint8) ([]byte, bool) {
	if !ok || len(buf) < 1 {
		return buf, false
	}
	*val = buf[0]
	return buf[1:], true
}

func DecodeUint8s(buf []byte, ok bool, vals ...*uint8) ([]byte, bool) {
	if !ok || len(buf) < len(vals) {
		return buf, false
	}
	for _, val := range vals {
		*val = buf[0]
		buf = buf[1:]
	}
	return buf, true
}

func DecodeUint16(buf []byte, ok bool, val *uint16) ([]byte, bool) {
	if !ok || len(buf) < 2 {
		return buf, false
	}
	*val = binary.LittleEndian.Uint16(buf)
	return buf[2:], true
}

func DecodeUint16s(buf []byte, ok bool, vals ...*uint16) ([]byte, bool) {
	if !ok || len(buf) < 2*len(vals) {
		return buf, false
	}
	for _, val := range vals {
		*val = binary.LittleEndian.Uint16(buf)
		buf = buf[2:]
	}
	return buf, true
}

func DecodeInt16s(buf []byte, ok bool, vals ...*int16) ([]byte, bool) {
	if !ok || len(buf) < 2*len(vals) {
		return buf, false
	}
	for _, val := range vals {
		*val = int16(binary.LittleEndian.Uint16(buf))
		buf = buf[2:]
	}
	return buf, true
}

func DecodeUint32(buf []byte, ok bool, val *uint32) ([]byte, bool) {
	if !ok || len(buf) < 4 {
		return buf, false
	}
	*val = binary.LittleEndian.Uint32(buf)
	return buf[4:], true
}

func DecodeUint32s(buf []byte, ok bool, vals ...*uint32) ([]byte, bool) {
	if !ok || len(buf) < len(vals)*4 {
		return buf, false
	}
	for _, val := range vals {
		*val = binary.LittleEndian.Uint32(buf)
		buf = buf[4:]
	}
	return buf, true
}

func DecodeInt32s(buf []byte, ok bool, vals ...*int32) ([]byte, bool) {
	if !ok || len(buf) < len(vals)*4 {
		return buf, false
	}
	for _, val := range vals {
		*val = int32(binary.LittleEndian.Uint32(buf))
		buf = buf[4:]
	}
	return buf, true
}

func DecodeSkip(buf []byte, ok bool, n int) ([]byte, bool) {
	if !ok || len(buf) < n {
		return buf, false
	}
	return buf[n:], true
}

func DecodeString(buf []byte, ok bool, val *string) ([]byte, bool) {
	if !ok || len(buf) < 4 {
		return buf, false
	}
	size := binary.LittleEndian.Uint32(buf)
	if len(buf) < int(size) {
		return buf, false
	}
	pad := size % 4
	if pad != 0 {
		pad = 4 - pad
	}
	buf = buf[4:]
	if size&0x80000000 == 0 {
		// UTF-8/ASCII
		*val = string(buf[:size])
		return buf[size+pad:], true
	}
	// UTF-16
	size &= 0x7fffffff
	size16 := size / 2
	buf16 := make([]uint16, size16)
	for i := 0; i < int(size16); i++ {
		buf16[i] = binary.LittleEndian.Uint16(buf)
		buf = buf[2:]
	}
	*val = string(utf16.Decode(buf16))
	return buf, true
}

func IsError(mcode MsgCode) error {
	switch mcode {
	case MsgInvalidRequest:
		return ErrInvalidRequest
	case MsgBadParameter:
		return ErrBadParameter
	case MsgBadLogin:
		return ErrBadLogin
	}
	return nil
}

func ValidateResponse(m Message, id uint16, code MsgCode) error {
	if len(m) < slnet.MessageHeaderSize {
		return fmt.Errorf("message too small: (%v < %v): %w", len(m), slnet.MessageHeaderSize, ErrInvalidResponse)
	}
	mcode := m.Code()
	if mcode == code+1 {
		return nil
	}
	if m.ID() != id {
		return fmt.Errorf("unexpected message id (%v != %v): %w", m.ID(), id, ErrUnexpectedResponseID)
	}
	if err := IsError(mcode); err != nil {
		return err
	}
	return fmt.Errorf("unexpected msg code (%v != %v): %w", mcode, code, ErrUnexpectedResponseCode)
}

func IsResponse(m Message, id uint16, code MsgCode) (bool, error) {
	if len(m) < slnet.MessageHeaderSize {
		return false, fmt.Errorf("message too small: (%v < %v): %w", len(m), slnet.MessageHeaderSize, ErrInvalidResponse)
	}
	mcode := m.Code()
	if err := IsError(mcode); err != nil {
		return false, err
	}
	return (mcode == code+1) && (m.ID() == id), nil
}

func DecodeDateTime(m Message) (time.Time, error) {
	pl := m.Payload()
	if g, w := len(pl), 9*2; g < w {
		return time.Time{}, fmt.Errorf("DecodeDateTime: payload too small: (%v < %v): %w", g, w, ErrInvalidResponse)
	}

	var year, month, unused, day, hour, minute, second, millisecond, autoDST uint16

	DecodeUint16s(pl, true, &year, &month, &unused, &day, &hour, &minute, &second, &millisecond, &autoDST)

	// Ignore the autoDST value for now.
	_ = autoDST

	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), int(millisecond)*1_000_000, time.UTC), nil
}

func DecodeVersion(m Message) string {
	var v string
	DecodeString(m.Payload(), true, &v)
	return v
}

func send(ctx context.Context, s Session, m Message, maxRetries int) error {
	var err error
	for i := range maxRetries {
		s.Send(ctx, m)
		err = s.Err()
		if err == nil {
			return nil
		}
		if i < maxRetries-1 {
			Logger(ctx).Log(ctx, slog.LevelInfo, "retrying", "op", m.Code(), "id", m.ID(), "err", err)
		}
	}
	return err
}

func sendAndValidate(ctx context.Context, s Session, m Message, id uint16, code MsgCode) (Message, error) {
	maxRetries := 3
	if err := send(ctx, s, m, 3); err != nil {
		return nil, err
	}
	for range maxRetries {
		msg := s.ReadUntil(ctx)
		if err := s.Err(); err != nil {
			return nil, err
		}
		rm := Message(msg)
		rm.SetID(id)
		if err := s.Err(); err != nil {
			return nil, err
		}
		ok, err := IsResponse(rm, id, code)
		if err != nil {
			return rm, err
		}
		if !ok {
			Logger(ctx).Log(ctx, slog.LevelInfo, "retrying", "expected_code", code, "expected_id", id, "actual_code", rm.Code(), "actual_id", rm.ID())
			continue
		}
		if err := ValidateResponse(rm, id, code); err != nil {
			return nil, err
		}
		return rm, nil
	}
	return nil, ErrNoValidResponse
}
