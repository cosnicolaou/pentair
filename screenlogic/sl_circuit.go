// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"context"

	"cloudeng.io/logging/ctxlog"
	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/pentair/screenlogic/protocol"
)

type CircuitConfig struct {
	ID int `yaml:"id"`
}

func NewCircuit(_ devices.Options) *Circuit {
	c := &Circuit{}
	return c
}

type Circuit struct {
	devices.DeviceBase[CircuitConfig]

	adapter *Adapter
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

var circuitState = map[bool]string{
	true:  "on",
	false: "off",
}

func (c *Circuit) setState(ctx context.Context, state bool) (any, error) {
	ctx, sess, err := c.adapter.session(ctx)
	if err != nil {
		return nil, err
	}
	defer sess.Release()

	circuit := c.DeviceConfigCustom.ID
	err = protocol.SetCircuitState(ctx, sess, circuit, state)
	if err != nil {
		ctxlog.Error(ctx, "screenlogic: failed to set circuit state", "op", circuitState[state], "circuit", circuit, "err", err)
		return nil, err
	}
	ctxlog.Info(ctx, "screenlogic: circuit state set", "op", circuitState[state], "circuit", circuit)
	return nil, err
}

func (c *Circuit) On(ctx context.Context, _ devices.OperationArgs) (any, error) {
	return c.setState(ctx, true)
}

func (c *Circuit) Off(ctx context.Context, _ devices.OperationArgs) (any, error) {
	return c.setState(ctx, false)
}
