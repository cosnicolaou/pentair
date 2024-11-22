// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package protocol

import "fmt"

type MsgID int
type MsgCode uint16

const (
	MsgLocalLogin     MsgCode = 27
	MsgBadLogin       MsgCode = 13
	MsgInvalidRequest MsgCode = 30
	MsgBadParameter   MsgCode = 31

	MsgGetDateTime MsgCode = 8110
	MsgGetVersion  MsgCode = 8120
	MsgGetConfig   MsgCode = 12532
	MsgGetStatus   MsgCode = 12526

	MsgButtonPress MsgCode = 12530
)

var (
	ErrBadLogin               = fmt.Errorf("bad login")
	ErrUnexpectedResponseID   = fmt.Errorf("unexpected response ID")
	ErrUnexpectedResponseCode = fmt.Errorf("unexpected response code")
	ErrInvalidRequest         = fmt.Errorf("invalid request")
	ErrInvalidResponse        = fmt.Errorf("invalid response")
	ErrBadParameter           = fmt.Errorf("bad parameter")
)

type ControllerState int

const (
	ControllerUnknownState ControllerState = iota
	ControllerReady
	ControllerSync
	ControllerService
)

type EquipmentFlags int

const (
	Solar EquipmentFlags = 1 << iota
	Solar_Heat_Pump
	Chlorinator
	IntelliBright
	IntelliFlo_0
	IntelliFlo_1
	IntelliFlo_2
	IntelliFlo_3
	IntelliFlo_4
	IntelliFlo_5
	IntelliFlo_6
	IntelliFlo_7
	NoSpecialLights
	Cooling
	MagicStream
	IntelliChem
	HybridHeater
)

func (ef EquipmentFlags) hasIntelliFlo(idx int) bool {
	return ef&(1<<uint(idx+4)) != 0
}

type ColorMode int

const (
	ColorAllOff ColorMode = iota
	ColorAllOn
	ColorSet
	ColorSync
	ColorSwim
	ColorParty
	ColorRomance
	ColorCaribbean
	ColorAmerican
	ColorSunset
	ColorRoyal
	ColorSave
	ColorRecall
	ColorBlue
	ColorGreen
	ColorRed
	ColorMagenta
	ColorThumper
	ColorNextMode
	ColorReset
	ColorHold
)

type CircuitFunction int

const (
	CircuitGeneric CircuitFunction = iota
	CircuitSpa
	CircuitPool
	CircuitSecondSpa
	CircuitSecondPool
	CircuitMasterCleaner
	CircuitCleaner
	CircuitLight
	CircuitDimmer
	CircuitSAMLight
	CircuitSALLight
	CircuitPhotoNextGen
	CircuitColorWheel
	CircuitValve
	CircuitSpillway
	CircuitFloorCleaner
	CircuitIntelliBrite
	CircuitMagicStream
	CircuitDimmer25
)

func (cf CircuitFunction) String() string {
	switch cf {
	case CircuitGeneric:
		return "Generic"
	case CircuitSpa:
		return "Spa"
	case CircuitPool:
		return "Pool"
	case CircuitSecondSpa:
		return "Second Spa"
	case CircuitSecondPool:
		return "Second Pool"
	case CircuitMasterCleaner:
		return "Master Cleaner"
	case CircuitCleaner:
		return "Cleaner"
	case CircuitLight:
		return "Light"
	case CircuitDimmer:
		return "Dimmer"
	case CircuitSAMLight:
		return "SAM Light"
	case CircuitSALLight:
		return "SAL Light"
	case CircuitPhotoNextGen:
		return "Photo Next Gen"
	case CircuitColorWheel:
		return "Color Wheel"
	case CircuitValve:
		return "Valve"
	case CircuitSpillway:
		return "Spillway"
	case CircuitFloorCleaner:
		return "Floor Cleaner"
	case CircuitIntelliBrite:
		return "IntelliBrite"
	case CircuitMagicStream:
		return "Magic Stream"
	case CircuitDimmer25:
		return "Dimmer 25"
	}
	return ""
}

type CircuitInterface int

const (
	InterfacePool CircuitInterface = iota
	InterfaceSpa
	InterfaceFeatures
	InterfaceSyncSwim
	InterfaceLights
	InterfaceDontShow
	InterfaceInvalid
)

func (ifc CircuitInterface) String() string {
	switch ifc {
	case InterfacePool:
		return "Pool"
	case InterfaceSpa:
		return "Spa"
	case InterfaceFeatures:
		return "Features"
	case InterfaceSyncSwim:
		return "Sync Swim"
	case InterfaceLights:
		return "Lights"
	case InterfaceDontShow:
		return "Don't Show"
	case InterfaceInvalid:
		return "Invalid"
	}
	return ""
}
