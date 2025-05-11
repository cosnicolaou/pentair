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
	"github.com/cosnicolaou/automation/net/streamconn"
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

	mgr      *streamconn.SessionManager
	ondemand *netutil.OnDemandConnection[streamconn.Transport, *Adapter]
}

func NewAdapter(_ devices.Options) *Adapter {
	pa := &Adapter{}
	pa.ondemand = netutil.NewOnDemandConnection(pa)
	pa.mgr = &streamconn.SessionManager{}
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

func (pa *Adapter) runOperation(ctx context.Context, op func(context.Context, *protocol.Session, devices.OperationArgs) (any, error), args devices.OperationArgs) (any, error) {
	ctx, sess, err := pa.session(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.Release()
	return op(ctx, sess, args)
}

func (pa *Adapter) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"gettime": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return pa.runOperation(ctx, pa.getTimeAndDate, args)
		},
		"getversion": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return pa.runOperation(ctx, pa.getVersion, args)
		},
		"getconfig": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return pa.runOperation(ctx, pa.getConfig, args)
		},
		"getstatus": func(ctx context.Context, args devices.OperationArgs) (any, error) {
			return pa.runOperation(ctx, pa.getStatus, args)
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

func (pa *Adapter) getTimeAndDate(ctx context.Context, sess *protocol.Session, args devices.OperationArgs) (any, error) {
	t, err := protocol.GetTimeAndDate(ctx, sess)
	if err == nil {
		fmt.Fprintf(args.Writer, "gettime: %v\n", t)
	}
	return struct {
		Time string `json:"time"`
	}{Time: t.String()}, err
}

func (pa *Adapter) getVersion(ctx context.Context, sess *protocol.Session, args devices.OperationArgs) (any, error) {
	version, err := protocol.GetVersionInfo(ctx, sess)
	if err == nil {
		fmt.Fprintf(args.Writer, "version: %v\n", version)
	}
	return struct {
		Version string `json:"version"`
	}{Version: version}, err
}

func (pa *Adapter) getConfig(ctx context.Context, sess *protocol.Session, args devices.OperationArgs) (any, error) {
	cfg, err := protocol.GetControllerConfig(ctx, sess)
	if err == nil {
		pa.FormatConfig(args.Writer, cfg)
	}
	return cfg, err
}

func (pa *Adapter) getStatus(ctx context.Context, sess *protocol.Session, args devices.OperationArgs) (any, error) {
	status, err := protocol.GetControllerStatus(ctx, sess)
	if err == nil {
		pa.FormatStatus(args.Writer, status)
	}
	return status, err
}

func (pa *Adapter) Connect(ctx context.Context, idle netutil.IdleReset) (streamconn.Transport, error) {
	ctxlog.Info(ctx, "screenlogic: connect: dialing", "ip", pa.ControllerConfigCustom.IPAddress)
	conn, err := slnet.Dial(ctx, pa.ControllerConfigCustom.IPAddress, pa.Timeout)
	if err != nil {
		return nil, err
	}
	session := protocol.NewSession(pa.mgr.New(conn, idle))
	defer session.Release()
	// Connect, there is no authentication for the screenlogic adapters
	// on a local network.
	ctxlog.Info(ctx, "screenlogic: connect: logging in", "ip", pa.ControllerConfigCustom.IPAddress)
	if err := protocol.Login(ctx, session); err != nil {
		conn.Close(ctx)
		return nil, err
	}
	ctxlog.Info(ctx, "screenlogic: connect: logged in", "ip", pa.ControllerConfigCustom.IPAddress)
	return conn, nil
}

func (pa *Adapter) Disconnect(ctx context.Context, conn streamconn.Transport) error {
	return conn.Close(ctx)
}

func (pa *Adapter) session(ctx context.Context) (context.Context, *protocol.Session, error) {
	ctx = ctxlog.WithAttributes(ctx, "protocol", "screenlogic")
	conn, idle, err := pa.ondemand.Connection(ctx)
	if err != nil {
		return ctx, nil, err
	}
	session := protocol.NewSession(pa.mgr.New(conn, idle))
	return ctx, session, nil
}

func (pa *Adapter) Close(ctx context.Context) error {
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
