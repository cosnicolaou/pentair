// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package screenlogic

import (
	"fmt"

	"github.com/cosnicolaou/automation/devices"
)

func SupportedDevices() devices.SupportedDevices {
	return devices.SupportedDevices{
		"circuit": func(string, devices.Options) (devices.Device, error) { return nil, nil },
	}
}

func SupportedControllers() devices.SupportedControllers {
	return devices.SupportedControllers{
		"screenlogic-adapter": NewController,
	}
}

func NewController(typ string, opts devices.Options) (devices.Controller, error) {
	switch typ {
	case "screenlogic-adapter":
		return NewSLAdapter(opts), nil
	}
	return nil, fmt.Errorf("unsupported pentair screenlogic type %s", typ)
}
