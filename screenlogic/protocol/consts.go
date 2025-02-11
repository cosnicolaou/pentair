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
	ErrNoValidResponse        = fmt.Errorf("no valid response received")
)

type ControllerState int

const (
	ControllerUnknownState ControllerState = iota
	ControllerReady
	ControllerSync
	ControllerService
)

func (cs ControllerState) String() string {
	switch cs {
	case ControllerUnknownState:
		return "Unknown"
	case ControllerReady:
		return "Ready"
	case ControllerSync:
		return "Sync"
	case ControllerService:
		return "Service"
	}
	return ""
}

type EquipmentFlags int

const (
	Solar EquipmentFlags = 1 << iota
	SolarHeatPump
	Chlorinator
	IntelliBright
	IntelliFlo0
	IntelliFlo1
	IntelliFlo2
	IntelliFlo3
	IntelliFlo4
	IntelliFlo5
	IntelliFlo6
	IntelliFlo7
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

var (
	cfLookup = []string{
		"Generic",
		"Spa",
		"Pool",
		"Second Spa",
		"Second Pool",
		"Master Cleaner",
		"Cleaner",
		"Light",
		"Dimmer",
		"SAM Light",
		"SAL Light",
		"Photo Next Gen",
		"Color Wheel",
		"Valve",
		"Spillway",
		"Floor Cleaner",
		"IntelliBrite",
		"Magic Stream",
		"Dimmer 25",
	}
)

func (cf CircuitFunction) String() string {
	if cf >= 0 && int(cf) < len(cfLookup) {
		return cfLookup[cf]
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

var (
	ciLookup = []string{
		"Pool",
		"Spa",
		"Features",
		"Sync Swim",
		"Lights",
		"Don't Show",
		"Invalid",
	}
)

func (ifc CircuitInterface) String() string {
	if ifc >= 0 && int(ifc) < len(ciLookup) {
		return ciLookup[ifc]
	}
	return ""
}
