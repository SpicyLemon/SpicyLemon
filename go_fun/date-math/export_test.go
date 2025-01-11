package main

import "io"

// This file exposes some private stuff so for the purpose of unit tests.

var (
	// ProcessFlags is a test-only exposure of processFlags.
	ProcessFlags = processFlags
	// IsPipeInd is a test-only exposure of isPipeInd.
	IsPipeInd = isPipeInd
	// MainE is a test-only exposure of mainE.
	MainE = mainE

	// MakeNamedFormat is a test-only exposure of makeNamedFormat.
	MakeNamedFormat = makeNamedFormat

	// SetOutputFormatByName is a test-only exposure of setOutputFormatByName.
	SetOutputFormatByName = setOutputFormatByName
	// SetOutputFormatByValue is a test-only exposure of setOutputFormatByValue.
	SetOutputFormatByValue = setOutputFormatByValue
	// SetInputFormatByName is a test-only exposure of setInputFormatByName.
	SetInputFormatByName = setInputFormatByName
	// SetInputFormatByValue is a test-only exposure of setInputFormatByValue
	SetInputFormatByValue = setInputFormatByValue
)

// CalcArgs is a test-only exposure of the calcArgs type.
type CalcArgs = calcArgs

// GetArgs is a test-only exposure of getArgs.
func GetArgs(argsIn []string, stdout io.Writer) (*CalcArgs, bool, error) {
	rv, ok, err := getArgs(argsIn, stdout)
	return rv, ok, err
}

// CombineArgs is a test-only exposure of combineArgs.
func CombineArgs(argsIn []string) *CalcArgs {
	return combineArgs(argsIn)
}
