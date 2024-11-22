// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/automation/net/netutil"
	"github.com/cosnicolaou/automation/net/streamconn"
	"github.com/cosnicolaou/pentair/screenlogic/protocol"
	"github.com/cosnicolaou/pentair/screenlogic/slnet"
	"gopkg.in/yaml.v3"
)

type AdapterConfig struct {
	IPAddress string        `yaml:"ip_address"`
	Timeout   time.Duration `yaml:"timeout"`
	KeepAlive time.Duration `yaml:"keep_alive"`
}

type Adapter struct {
	devices.ControllerConfigCommon
	AdapterConfig `yaml:",inline"`
	logger        *slog.Logger

	mu      sync.Mutex
	session protocol.Session
	manager *streamconn.Manager
}

func NewSLAdapter(opts devices.Options) *Adapter {
	return &Adapter{
		logger:  opts.Logger.With("protocol", "screenlogic"),
		manager: streamconn.NewManager(),
	}
}

func (pa *Adapter) SetConfig(c devices.ControllerConfigCommon) {
	pa.ControllerConfigCommon = c
}

func (pa *Adapter) Config() devices.ControllerConfigCommon {
	return pa.ControllerConfigCommon
}

func (pa *Adapter) CustomConfig() any {
	return pa.AdapterConfig
}

func (pa *Adapter) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&pa.AdapterConfig); err != nil {
		return err
	}
	return nil
}

func (pa *Adapter) Implementation() any {
	return pa
}

func (pa *Adapter) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"gettime": func(ctx context.Context, args devices.OperationArgs) error {
			t, err := pa.GetTime(ctx)
			if err == nil {
				fmt.Fprintf(args.Writer, "gettime: %v\n", t)
			}
			return err
		},
		"getversion": func(ctx context.Context, args devices.OperationArgs) error {
			version, err := protocol.GetVersionInfo(ctx, pa.session)
			if err == nil {
				fmt.Fprintf(args.Writer, "version: %v\n", version)
			}
			return err
		},
		"getconfig": func(ctx context.Context, args devices.OperationArgs) error {
			cfg, err := protocol.GetControllerConfig(ctx, pa.session)
			if err == nil {
				fmt.Fprintf(args.Writer, "config: %+v\n", cfg)
			}
			return err
		},
		"getstatus": func(ctx context.Context, args devices.OperationArgs) error {
			status, err := protocol.GetControllerStatus(ctx, pa.session)
			if err == nil {
				fmt.Fprintf(args.Writer, "status: %+v\n", status)
			}
			return err
		},
	}
}

func (pa *Adapter) OperationsHelp() map[string]string {
	return map[string]string{
		"gettime":    "get the current time, date and timezone",
		"getconfig":  "get the current system configuration",
		"getstatus":  "get the current system satus",
		"getversion": "get the adapter version",
	}
}

func (pa *Adapter) GetTime(ctx context.Context) (time.Time, error) {
	s := pa.Session(ctx)
	return protocol.GetTimeAndDate(ctx, s)
}

func (pa *Adapter) Session(ctx context.Context) protocol.Session {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	if sess := pa.manager.Session(); sess != nil {
		return sess.(protocol.Session)
	}
	transport, err := slnet.Dial(ctx, pa.IPAddress, pa.Timeout, pa.logger)
	if err != nil {
		return protocol.NewErrorSession(err)
	}
	idle := netutil.NewIdleTimer(pa.KeepAlive)
	session := protocol.NewSession(transport, idle)

	// Connect.
	if err := protocol.Login(ctx, session); err != nil {
		session.Close(ctx)
		return protocol.NewErrorSession(err)
	}
	pa.manager.ManageSession(session, idle)
	return session
}

func (pa *Adapter) Close(ctx context.Context) error {
	return pa.manager.Close(ctx, time.Minute)
}
