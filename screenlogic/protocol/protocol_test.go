// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/cosnicolaou/pentair/screenlogic/protocol"
)

func TestMessage(t *testing.T) {
	m := protocol.NewMessage(1, 2, []byte{'a', 'b', 'c'})
	if got, want := m.ID(), uint16(1); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := m.Code(), protocol.MsgCode(2); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := m.Size(), uint32(3); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := m.Payload(), []byte{'a', 'b', 'c'}; !bytes.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	m = protocol.NewEmptyMessage(1, 2, 3)
	if got, want := m.ID(), uint16(1); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := m.Code(), protocol.MsgCode(2); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := m.Size(), uint32(3); got != want {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got, want := m.Payload(), []byte{0, 0, 0}; !bytes.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	for _, i := range []int{1, 2, 3, 4, 5, 6, 7, 8, 9} {
		r := i % 4
		if r != 0 {
			r = 4 - r
		}
		fmt.Printf("i: %d, r: %d\n", i, i+r)
	}
	t.Fail()
}
