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
	"github.com/cosnicolaou/automation/net/streamconn"
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

	mu           sync.Mutex
	idle         *streamconn.IdleTimer
	session      streamconn.Session
	closeContext context.Context
	closeCancel  context.CancelFunc
	closeCh      chan struct{}
}

func NewAdapter(opts devices.Options) *Adapter {
	return &Adapter{
		logger: opts.Logger,
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
	return p
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
	}
}

func (pa *Adapter) OperationsHelp() map[string]string {
	return map[string]string{
		"gettime":     "get the current time, date and timezone",
		"getlocation": "get the current location in latitude and longitude",
		"getsuntimes": "get the current sunrise and sunset times in local time",
		"os_version":  "get the OS version running on QS processor",
	}
}

func (pa *Adapter) GetTime(ctx context.Context) (time.Time, error) {

	return time.Now(), nil
}
