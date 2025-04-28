// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloudeng.io/logging/ctxlog"
	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/automation/net/netutil"
	"github.com/cosnicolaou/pentair/screenlogic/protocol"
	"github.com/cosnicolaou/pentair/screenlogic/slnet"
	"gopkg.in/yaml.v3"
)

type AdapterConfig struct {
	IPAddress string        `yaml:"ip_address"`
	KeepAlive time.Duration `yaml:"keep_alive"`
}

type Adapter struct {
	devices.ControllerBase[AdapterConfig]

	ondemand *netutil.OnDemandConnection[protocol.Session, *Adapter]
}

func NewAdapter(opts devices.Options) *Adapter {
	pa := &Adapter{}
	pa.ondemand = netutil.NewOnDemandConnection(pa, protocol.NewErrorSession)
	return pa
}

func (pa *Adapter) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&pa.ControllerConfigCustom); err != nil {
		return err
	}
	if pa.ControllerConfigCustom.KeepAlive == 0 {
		return fmt.Errorf("keep_alive must be specified")
	}
	pa.ondemand.SetKeepAlive(pa.ControllerConfigCustom.KeepAlive)
	return nil
}

func (pa *Adapter) Implementation() any {
	return pa
}

func (pa *Adapter) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"gettime": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := pa.Session(ctx)
			t, err := protocol.GetTimeAndDate(ctx, sess)
			if err == nil {
				fmt.Fprintf(args.Writer, "gettime: %v\n", t)
			}
			return struct {
				Time string `json:"time"`
			}{Time: t.String()}, err
		},
		"getversion": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := pa.Session(ctx)
			version, err := protocol.GetVersionInfo(ctx, sess)
			if err == nil {
				fmt.Fprintf(args.Writer, "version: %v\n", version)
			}
			return struct {
				Version string `json:"version"`
			}{Version: version}, err
		},
		"getconfig": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := pa.Session(ctx)
			cfg, err := protocol.GetControllerConfig(ctx, sess)
			if err == nil {
				pa.FormatConfig(args.Writer, cfg)
			}
			return cfg, err
		},
		"getstatus": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			ctx, sess := pa.Session(ctx)
			status, err := protocol.GetControllerStatus(ctx, sess)
			if err == nil {
				pa.FormatStatus(args.Writer, status)
			}
			return status, err
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

func (pa *Adapter) Connect(ctx context.Context, idle netutil.IdleReset) (protocol.Session, error) {
	ctxlog.Info(ctx, "screenlogic: connect: dialing", "ip", pa.ControllerConfigCustom.IPAddress)
	transport, err := slnet.Dial(ctx, pa.ControllerConfigCustom.IPAddress, pa.Timeout)
	if err != nil {
		return nil, err
	}
	session := protocol.NewSession(transport, idle)
	// Connect, there is no authentication for the screenlogic adapters
	// on a local network.
	ctxlog.Info(ctx, "screenlogic: connect: logging in", "ip", pa.ControllerConfigCustom.IPAddress)
	if err := protocol.Login(ctx, session); err != nil {
		session.Close(ctx)
		return nil, err
	}
	ctxlog.Info(ctx, "screenlogic: connect: logged in", "ip", pa.ControllerConfigCustom.IPAddress)
	return session, nil
}

func (pa *Adapter) Disconnect(ctx context.Context, sess protocol.Session) error {
	return sess.Close(ctx)
}

func (pa *Adapter) loggingContext(ctx context.Context) context.Context {
	return ctxlog.WithAttributes(ctx, "protocol", "screenlogic")
}

func (pa *Adapter) Session(ctx context.Context) (context.Context, protocol.Session) {
	ctx = pa.loggingContext(ctx)
	return ctx, pa.ondemand.Connection(ctx)
}

func (pa *Adapter) Close(ctx context.Context) error {
	ctx = pa.loggingContext(ctx)
	return pa.ondemand.Close(ctx)
}

func (pa *Adapter) FormatConfig(out io.Writer, cfg protocol.ControllerConfig) {
	if out == nil {
		return
	}
	fmt.Fprintf(out, "Address  : %v\n", pa.ControllerConfigCustom.IPAddress)
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
	fmt.Fprintf(out, "Address  : %v\n", pa.ControllerConfigCustom.IPAddress)
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
