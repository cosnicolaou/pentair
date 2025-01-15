// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"context"
	"log/slog"

	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/pentair/screenlogic/protocol"
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
	devices.DeviceBase[CircuitConfig]
	adapter *Adapter
	logger  *slog.Logger
}

func (c *Circuit) SetController(ctrl devices.Controller) {
	c.adapter = ctrl.Implementation().(*Adapter)
}

func (c *Circuit) ControlledBy() devices.Controller {
	return c.adapter
}

func (c *Circuit) OperationsHelp() map[string]string {
	return map[string]string{
		"on":  "turn the circuit on",
		"off": "turn the circuit off",
	}
}

func (c *Circuit) Operations() map[string]devices.Operation {
	return map[string]devices.Operation{
		"on":  c.On,
		"off": c.Off,
	}
}

func (c *Circuit) On(ctx context.Context, _ devices.OperationArgs) error {
	sess := c.adapter.Session(ctx)
	return protocol.SetCircuitState(ctx, sess, c.DeviceConfigCustom.ID, true)
}

func (c *Circuit) Off(ctx context.Context, _ devices.OperationArgs) error {
	sess := c.adapter.Session(ctx)
	return protocol.SetCircuitState(ctx, sess, c.DeviceConfigCustom.ID, false)
}
