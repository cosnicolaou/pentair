// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import (
	"context"
	"fmt"
)

type hardwareType struct {
	hw   uint8
	name string
}

var controllerTypes = [][]hardwareType{
	{{0, "IntelliTouch i5+3S"}},                            // 0
	{{0, "IntelliTouch i7+3"}},                             // 1
	{{0, "IntelliTouch i9+3"}},                             // 2
	{{0, "IntelliTouch i5+3S"}},                            // 3
	{{0, "IntelliTouch i9+3S"}},                            // 4
	{{0, "IntelliTouch i10+3D"}, {1, "IntelliTouch i10X"}}, // 5
	{{0, "IntelliTouch i10X"}},                             // 6
	{{}},                                                   // 7
	{{}},                                                   // 8
	{{}},                                                   // 9
	{{0, "SunTouch"}},                                      // 10
	{{0, "Suntouch/Intellicom"}},                           // 11
	{{}},                                                   // 12
	{
		{0, "EasyTouch2 8"}, // 13
		{1, "EasyTouch2 8P"},
		{2, "EasyTouch2 4"},
		{3, "EasyTouch2 4P"},
		{5, "EasyTouch2 PL4"},
		{6, "EasyTouch2 PSL4"},
	},
	{
		{0, "EasyTouch1 8"}, // 14
		{1, "EasyTouch1 8P"},
		{2, "EasyTouch1 4"},
		{3, "EasyTouch1 4P"},
	}}

func DecodeControllerHardware(controller, hardware uint8) (string, error) {
	if controller >= uint8(len(controllerTypes)) {
		return "", fmt.Errorf("Unknown controller type %d", controller)
	}
	if hardware >= uint8(len(controllerTypes[controller])) {
		return "", fmt.Errorf("Unknown hardware type %d", hardware)
	}
	return controllerTypes[controller][hardware].name, nil
}

func GetControllerConfig(ctx context.Context, s Session) (ControllerConfig, error) {
	id := s.NextID()
	m := NewEmptyMessage(id, MsgGetConfig, 8) // 2 INTs value 0.
	rm, err := sendAndValidate(ctx, s, m, id, MsgGetConfig)
	if err != nil {
		return ControllerConfig{}, fmt.Errorf("GetConfig: %w", err)
	}
	return DecodeControllerConfig(rm)
}

func DecodeControllerConfig(rm Message) (ControllerConfig, error) {
	var cfg ControllerConfig
	pl := rm.Payload()

	ok := true
	var id uint32
	pl, ok = DecodeUint32(pl, ok, &id)
	cfg.ID = int(id)
	pl, ok = DecodeSkip(pl, ok, 4) // Skip setpoint data.
	pl, ok = DecodeSkip(pl, ok, 1) // Skip celsius/faherenheit data.
	var cv, hw uint8
	pl, ok = DecodeUint8s(pl, ok, &cv, &hw)
	model, err := DecodeControllerHardware(cv, hw)
	if err != nil {
		return ControllerConfig{}, fmt.Errorf("DecodeControllerConfig: %w", err)
	}
	cfg.Model = model
	pl, ok = DecodeSkip(pl, ok, 1) // Skip controller data byte

	var flags uint32
	pl, ok = DecodeUint32(pl, ok, &flags)
	cfg.Equipment = EquipmentFlags(flags)

	var circuitName string
	pl, ok = DecodeString(pl, ok, &circuitName)
	var circuitCount uint32

	// Circuits
	pl, ok = DecodeUint32(pl, ok, &circuitCount)
	for range int(circuitCount) {
		var c Circuit
		var id uint32
		pl, ok = DecodeUint32(pl, ok, &id)
		c.ID = int(id)
		pl, ok = DecodeString(pl, ok, &c.Name)
		pl, ok = DecodeUint8(pl, ok, &c.Index)
		var fn, ifc, flags, colorSet, colorPos, colorStagger uint8
		pl, ok = DecodeUint8s(pl, ok, &fn, &ifc, &flags, &colorSet, &colorPos, &colorStagger)
		c.Function = CircuitFunction(fn)
		c.Interface = CircuitInterface(ifc)

		pl, ok = DecodeUint8(pl, ok, &c.DeviceID)

		cfg.Circuits = append(cfg.Circuits, c)

		var runtime uint16
		pl, ok = DecodeUint16(pl, ok, &runtime)
		pl, ok = DecodeSkip(pl, ok, 2)

		if !ok {
			break
		}
	}
	if !ok {
		return ControllerConfig{}, fmt.Errorf("DecodeControllerConfig: message too small: %w", ErrInvalidResponse)
	}

	// Colors
	var colorCount uint32
	pl, ok = DecodeUint32(pl, ok, &colorCount)
	for range int(colorCount) {
		var color string
		pl, ok = DecodeString(pl, ok, &color)
		var R, G, B uint32
		pl, ok = DecodeUint32s(pl, ok, &R, &G, &B)
		if !ok {
			break
		}
	}

	if !ok {
		return ControllerConfig{}, fmt.Errorf("DecodeControllerConfig: message too small: %w", ErrInvalidResponse)
	}

	pumpCount := 8
	for i := 0; i < pumpCount; i++ {
		var val uint8
		pl, ok = DecodeUint8(pl, ok, &val)
		if cfg.Equipment.hasIntelliFlo(i) {
			cfg.IntelliFlo = append(cfg.IntelliFlo, IntelliFlo{Value: val})
		}
	}

	if !ok {
		return ControllerConfig{}, fmt.Errorf("DecodeControllerConfig: message too small: %w", ErrInvalidResponse)
	}

	var flags2, alarms uint32
	pl, ok = DecodeUint32s(pl, ok, &flags2, &alarms)

	if !ok {
		return ControllerConfig{}, fmt.Errorf("DecodeControllerConfig: message too small: %w", ErrInvalidResponse)
	}

	if len(pl) > 0 {
		return ControllerConfig{}, fmt.Errorf("DecodeControllerConfig: spurious data: %w", ErrInvalidResponse)
	}

	return cfg, nil
}

type Circuit struct {
	ID        int
	Name      string
	Function  CircuitFunction
	Interface CircuitInterface
	Index     uint8
	DeviceID  uint8
}

type IntelliFlo struct {
	Value uint8
}

type ControllerConfig struct {
	Model      string
	ID         int
	Equipment  EquipmentFlags
	Circuits   []Circuit
	IntelliFlo []IntelliFlo
}

func GetControllerStatus(ctx context.Context, s Session) (ControllerStatus, error) {
	id := s.NextID()
	m := NewEmptyMessage(id, MsgGetStatus, 4) // 1 INT value 0.
	rm, err := sendAndValidate(ctx, s, m, id, MsgGetStatus)
	if err != nil {
		return ControllerStatus{}, fmt.Errorf("GetConfig: %w", err)
	}
	return DecodeControllerStatus(rm)
}

func DecodeControllerStatus(rm Message) (ControllerStatus, error) {
	var status ControllerStatus
	pl := rm.Payload()
	var state uint32
	ok := true
	pl, ok = DecodeUint32(pl, ok, &state)
	status.State = ControllerState(state)

	// Skip freeze mode, remotes, pool delay, spa delay and cleaner delay, + 3
	// unknown + air temp
	pl, ok = DecodeSkip(pl, ok, 5+3+4)

	var nBodies uint32
	pl, ok = DecodeUint32(pl, ok, &nBodies)
	for range int(nBodies) {
		// Skip body data, type (uint32)+ last temp, heat, heat set point, cool set point, heat mode (int32),
		pl, ok = DecodeSkip(pl, ok, 4+(5*4))
	}

	if !ok {
		return ControllerStatus{}, fmt.Errorf("DecodeControllerStatus: message too small: %w", ErrInvalidResponse)
	}

	var circuitCount uint32
	pl, ok = DecodeUint32(pl, ok, &circuitCount)
	for range int(circuitCount) {
		var c CircuitStatus
		var id, val uint32
		pl, ok = DecodeUint32(pl, ok, &id)
		pl, ok = DecodeUint32(pl, ok, &val)
		c.ID = int(id)
		c.State = val != 0
		var colorSet, colorPos, colorStagger, delay uint8
		pl, ok = DecodeUint8s(pl, ok, &colorSet, &colorPos, &colorStagger, &delay)
		status.Circuits = append(status.Circuits, c)
	}

	var pH, orp, saturation, saltPPM, pHTank, orpTank, alert int32
	pl, ok = DecodeInt32s(pl, ok, &pH, &orp, &saturation, &saltPPM, &pHTank, &orpTank, &alert)
	status.Alert = int(alert)

	if !ok {
		return ControllerStatus{}, fmt.Errorf("DecodeControllerStatus: message too small: %w", ErrInvalidResponse)
	}

	if len(pl) > 0 {
		return ControllerStatus{}, fmt.Errorf("DecodeControllerStatus: spurious data: %w", ErrInvalidResponse)
	}

	return status, nil
}

type CircuitStatus struct {
	ID    int
	State bool
}

type ControllerStatus struct {
	State    ControllerState
	Circuits []CircuitStatus
	Alert    int
}

func (cs ControllerStatus) StatusForID(id int) bool {
	for _, c := range cs.Circuits {
		if c.ID == id {
			return c.State
		}
	}
	return false
}

func (c ControllerConfig) CircuitName(id int) string {
	for _, c := range c.Circuits {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

func (c ControllerConfig) CircuitByID(id int) Circuit {
	for _, c := range c.Circuits {
		if c.ID == id {
			return c
		}
	}
	return Circuit{}
}

func (c ControllerConfig) CircuitBytName(name string) Circuit {
	for _, c := range c.Circuits {
		if c.Name == name {
			return c
		}
	}
	return Circuit{}
}

func SetCircuitState(ctx context.Context, s Session, circuitID int, state bool) error {
	m := NewEmptyMessage(0, MsgButtonPress, 3*4)
	pl := m.Payload()
	pl = AppendUint32(pl, 0)
	pl = AppendUint32(pl, uint32(circuitID))
	if state {
		AppendUint32(pl, 1)
	} else {
		AppendUint32(pl, 0)
	}
	id := s.NextID()
	rm, err := sendAndValidate(ctx, s, m, id, MsgButtonPress)
	if err != nil {
		return fmt.Errorf("SetCircuitState: %w", err)
	}
	if len(rm.Payload()) != 0 {
		return fmt.Errorf("SetCircuitState: unexpected response: %w", ErrInvalidResponse)
	}
	return nil
}
