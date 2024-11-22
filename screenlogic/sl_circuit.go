// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"context"
	"log/slog"
	"time"

	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/pentair/screenlogic/protocol"
	"gopkg.in/yaml.v3"
)

type CircuitConfig struct {
	ID int `yaml:"id"`
}

func NewCircuit(opts devices.Options) *Circuit {
	c := &Circuit{}
	c.logger = opts.Logger.With(
		"protocol", "screenlogic",
		"device", "circuit")
	return c
}

type Circuit struct {
	devices.DeviceConfigCommon
	CircuitConfig
	adapter *Adapter
	logger  *slog.Logger
}

func (c *Circuit) SetConfig(cfg devices.DeviceConfigCommon) {
	c.DeviceConfigCommon = cfg
}

func (c *Circuit) Config() devices.DeviceConfigCommon {
	return c.DeviceConfigCommon
}

func (c *Circuit) SetController(ctrl devices.Controller) {
	c.adapter = ctrl.Implementation().(*Adapter)
}

func (c *Circuit) ControlledByName() string {
	return c.Controller
}

func (c *Circuit) ControlledBy() devices.Controller {
	return c.adapter
}

func (c *Circuit) Timeout() time.Duration {
	return time.Minute
}

func (c *Circuit) CustomConfig() any {
	return c.CircuitConfig
}

func (c *Circuit) OperationsHelp() map[string]string {
	return map[string]string{
		"on":  "turn the circuit on",
		"off": "turn the circuit off",
	}
}

func (c *Circuit) UnmarshalYAML(node *yaml.Node) error {
	return node.Decode(&c.CircuitConfig)
}

func (c *Circuit) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"on":  c.On,
		"off": c.Off,
	}
}

func (c *Circuit) On(ctx context.Context, opts devices.OperationArgs) error {
	sess := c.adapter.Session(ctx)
	return protocol.SetCircuitState(ctx, sess, c.ID, true)
}

func (c *Circuit) Off(ctx context.Context, opts devices.OperationArgs) error {
	sess := c.adapter.Session(ctx)
	return protocol.SetCircuitState(ctx, sess, c.ID, false)
}
