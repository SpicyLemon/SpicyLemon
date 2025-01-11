package main

import (
	"errors"
	"fmt"
)

// Operation is a type for representing various operations that this can handle.
type Operation string

const (
	OpAdd Operation = "+"
	OpSub Operation = "-"
	OpMul Operation = "x" // Not * because lots of shells expand that.
	OpDiv Operation = "/"
)

// Validate returns an error if this operation isn't valid.
func (o Operation) Validate() error {
	if o != OpAdd && o != OpSub && o != OpMul && o != OpDiv {
		return fmt.Errorf("unknown operation %q: must be either %q or %q or %q or %q",
			string(o), OpAdd, OpSub, OpMul, OpDiv)
	}
	return nil
}

// Is returns true of the provided string represents this Operation.
func (o Operation) Is(arg string) bool {
	return string(o) == arg
}

// String converts this Operation into a string (and satisfies the fmt.Stringer interface).
func (o Operation) String() string {
	return string(o)
}

// Name returns the global variable name of this Operation.
func (o Operation) Name() string {
	switch o {
	case OpAdd:
		return "OpAdd"
	case OpSub:
		return "OpSub"
	case OpMul:
		return "OpMul"
	case OpDiv:
		return "OpDiv"
	}
	return fmt.Sprintf("Operation(%q)", string(o))
}

// IsOp returns true if the provided string is an Operation string.
func IsOp(arg string) bool {
	return Operation(arg).Validate() == nil
}

// ParseOperation attempts to convert a string into an Operation.
func ParseOperation(arg string) (Operation, error) {
	if len(arg) == 0 {
		return "", errors.New("empty operation argument not allowed")
	}
	o := Operation(arg)
	return o, o.Validate()
}
