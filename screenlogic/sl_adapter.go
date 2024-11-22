// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"context"
	"fmt"
	"io"
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

func NewAdapter(opts devices.Options) *Adapter {
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
			t, err := protocol.GetTimeAndDate(ctx, pa.Session(ctx))
			if err == nil {
				fmt.Fprintf(args.Writer, "gettime: %v\n", t)
			}
			return err
		},
		"getversion": func(ctx context.Context, args devices.OperationArgs) error {
			version, err := protocol.GetVersionInfo(ctx, pa.Session(ctx))
			if err == nil {
				fmt.Fprintf(args.Writer, "version: %v\n", version)
			}
			return err
		},
		"getconfig": func(ctx context.Context, args devices.OperationArgs) error {
			cfg, err := protocol.GetControllerConfig(ctx, pa.Session(ctx))
			if err == nil {
				pa.FormatConfig(args.Writer, cfg)
			}
			return err
		},
		"getstatus": func(ctx context.Context, args devices.OperationArgs) error {
			status, err := protocol.GetControllerStatus(ctx, pa.Session(ctx))
			if err == nil {
				pa.FormatStatus(args.Writer, status)
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

	// Connect, there is no actual authentication
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

func (pa *Adapter) FormatConfig(out io.Writer, cfg protocol.ControllerConfig) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, "Address  : %v\n", pa.IPAddress)
	fmt.Fprintf(out, "Model    : %v\n", cfg.Model)
	fmt.Fprintf(out, "ID       : %v\n", cfg.ID)
	fmt.Fprintf(out, "Circuits : #%v\n", len(cfg.Circuits))
	for _, c := range cfg.Circuits {
		fmt.Fprintf(out, "  % 5v : %10v", c.ID, c.Name)
		fmt.Fprintf(out, "  %20v % 20v\n", c.Function.String(), c.Interface.String())
	}
	fmt.Fprintf(out, "#Pumps   : %v\n", len(cfg.IntelliFlo))
	for i, p := range cfg.IntelliFlo {
		fmt.Fprintf(out, "  % 2v : %v\n", i, p.Value)
	}
}

func (pa *Adapter) FormatStatus(out io.Writer, st protocol.ControllerStatus) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, "Address  : %v\n", pa.IPAddress)
	fmt.Fprintf(out, "State    : %v\n", st.State.String())
	fmt.Fprintf(out, "Circuits : #%v\n", len(st.Circuits))
	for _, c := range st.Circuits {
		if c.State {
			fmt.Fprintf(out, "  % 5v : On\n", c.ID)
		} else {
			fmt.Fprintf(out, "  % 5v : Off\n", c.ID)
		}
	}
}
