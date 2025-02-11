// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/cosnicolaou/pentair/screenlogic/protocol"
	"github.com/cosnicolaou/pentair/screenlogic/slnet"
)

func exit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func onOff(state bool) string {
	if state {
		return "on"
	}
	return "off"
}

type idleReset struct{}

func (idleReset) Reset() {
}

func main() {
	ctx := context.Background()
	addr := "172.16.1.82:80"
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	conn, err := slnet.Dial(ctx, addr, 5*time.Second, logger)
	if err != nil {
		exit("failed to dial %s: %v\n", addr, err)
	}
	ctx = protocol.WithLogger(ctx, logger)
	sess := protocol.NewSession(conn, idleReset{})
	if err := protocol.Login(ctx, sess); err != nil {
		exit("failed to login to %s: %v\n", addr, err)
	}
	time, err := protocol.GetTimeAndDate(ctx, sess)
	if err != nil {
		exit("failed to connect to %s: %v\n", addr, err)
	}
	fmt.Printf("time: %s\n", time)

	version, err := protocol.GetVersionInfo(ctx, sess)
	if err != nil {
		exit("failed to get version info: %v\n", err)
	}
	fmt.Printf("version: %s\n", version)

	cfg, err := protocol.GetControllerConfig(ctx, sess)
	if err != nil {
		exit("failed to get equipment: %v\n", err)
	}
	fmt.Printf("config: %+v\n", cfg)

	status, err := protocol.GetControllerStatus(ctx, sess)
	if err != nil {
		exit("failed to get status: %v\n", err)
	}

	fmt.Printf("status: %+v\n", status)
	for _, cs := range status.Circuits {
		c := cfg.CircuitByID(cs.ID)
		fmt.Printf("%q: %v (%v: %v: %v)\n", c.Name, onOff(cs.State), c.ID, c.Function, c.Interface)
	}

	c := cfg.CircuitBytName("B. FEAT LT")
	if c.ID == 0 {
		exit("failed to find circuit\n")
	}
	fmt.Printf("ID: %d %v -> %v\n", c.ID, status.StatusForID(c.ID), !status.StatusForID(c.ID))
	if err := protocol.SetCircuitState(ctx, sess, c.ID, !status.StatusForID(c.ID)); err != nil {
		exit("failed to set circuit state: %v\n", err)
	}

	status, err = protocol.GetControllerStatus(ctx, sess)
	if err != nil {
		exit("failed to get status: %v\n", err)
	}

	fmt.Printf("status: %+v\n", status)
	for _, cs := range status.Circuits {
		c := cfg.CircuitByID(cs.ID)
		fmt.Printf("%q: %v (%v: %v: %v)\n", c.Name, onOff(cs.State), c.ID, c.Function, c.Interface)
	}
}
