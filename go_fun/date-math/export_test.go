package main

// This file exposes some private stuff so for the purpose of unit tests.

var (
	// GetArgs is a test-only exposure of getArgs.
	GetArgs = getArgs
	// ProcessFlags is a test-only exposure of processFlags.
	ProcessFlags = processFlags
	// CombineArgs is a test-only exposure of combineArgs.
	CombineArgs = combineArgs
	// GetNextValueArg is a test-only exposure of getNextValueArg.
	GetNextValueArg = getNextValueArg
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
