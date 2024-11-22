// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package slnet_test

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/cosnicolaou/pentair/screenlogic/protocol"
	"github.com/cosnicolaou/pentair/screenlogic/slnet"
)

type screenlogicGateway struct {
	net.Conn
}

func newListener(t *testing.T) net.Listener {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	return listener
}

func runGateway(listener net.Listener, errCh chan error) {
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			return
		}
		gw := &screenlogicGateway{conn}
		errCh <- gw.run()
	}()
}

//go:embed testdata/msg_*
var embeddedMessages embed.FS

var (
	cannedResponses = map[uint32][]byte{}
)

func init() {
	msgs, err := embeddedMessages.ReadDir("testdata")
	if err != nil {
		panic(err)
	}
	for _, msg := range msgs {
		var id uint32
		_, err := fmt.Sscanf(msg.Name(), "msg_%d", &id)
		if err != nil {
			panic(err)
		}
		b, err := embeddedMessages.ReadFile(fmt.Sprintf("testdata/%s", msg.Name()))
		if err != nil {
			panic(err)
		}
		cannedResponses[id] = b
	}
}

func (slg *screenlogicGateway) run() error {
	for {
		buf := make([]byte, 4096)
		n, err := slg.Read(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		m := protocol.Message(buf[:n])
		slg.Write(m)
	}
}

func TestSLConn(t *testing.T) {
	ctx := context.Background()
	errCh := make(chan error, 1)
	gl := newListener(t)
	runGateway(gl, errCh)

	logRecorder := bytes.NewBuffer(nil)
	logger := slog.New(slog.NewJSONHandler(logRecorder, nil))

	conn, err := slnet.Dial(ctx, gl.Addr().String(), time.Minute, logger)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	_ = conn
	fmt.Printf("close....\n")
}
