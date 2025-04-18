// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"context"
	"time"

	"cloudeng.io/logging/ctxlog"
	"github.com/cosnicolaou/automation/devices"
	"github.com/cosnicolaou/pentair/screenlogic/protocol"
)

type CircuitConfig struct {
	ID int `yaml:"id"`
}

func NewCircuit(opts devices.Options) *Circuit {
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

func (c *Circuit) setState(ctx context.Context, state bool) (any, error) {
	ctx = c.adapter.loggingContext(ctx)
	maxRetries := c.DeviceConfigCommon.RetryConfig.Retries
	s := "SetCircuiteState off"
	if state {
		s = "SetCircuiteState on"
	}
	var lastErr error
	for i := range maxRetries {
		sess := c.adapter.Session(ctx)
		err := protocol.SetCircuitState(ctx, sess, c.DeviceConfigCustom.ID, state)
		if err == nil {
			ctxlog.Info(ctx, "screenlogic: set circuit state", "op", s, "id", c.DeviceConfigCustom.ID)
			return nil, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			ctxlog.Info(ctx, "screenlogic: retrying", "retries", i, "max_retries", maxRetries, "op", s, "id", c.DeviceConfigCustom.ID, "err", err)
			time.Sleep(c.DeviceConfigCommon.RetryConfig.Timeout)
			continue
		}
		break
	}
	ctxlog.Error(ctx, "screenlogic: failed to set circuit state", "op", s, "id", c.DeviceConfigCustom.ID, "err", lastErr)
	return nil, lastErr
}

func (c *Circuit) On(ctx context.Context, _ devices.OperationArgs) (any, error) {
	return c.setState(ctx, true)
}

func (c *Circuit) Off(ctx context.Context, _ devices.OperationArgs) (any, error) {
	return c.setState(ctx, false)
}
