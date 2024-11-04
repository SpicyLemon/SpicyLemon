package main

import (
	"errors"
	"fmt"
	"time"
)

// DoCalculation processes the provided args as a formula and returns the result.
func DoCalculation(formula []string) (*DTVal, error) {
	CurStep = 0
	if len(formula) == 0 {
		return nil, errors.New("no formula provided")
	}

	rv, err := ParseDTVal(formula[0])
	if err != nil {
		return nil, err
	}
	verboseStepf(stepValue, "%s  <= %q", rv, formula[0])

	var op Operation
	var val *DTVal
	for i := 1; i < len(formula); {
		CurStep++
		op, err = ParseOperation(formula[i])
		if err != nil {
			return nil, err
		}
		verboseStepf(stepOp, "%s  <= %q", op, formula[i])
		i++
		if i >= len(formula) {
			return nil, fmt.Errorf("formula ends with operation %q: must end in value", op)
		}

		val, err = ParseDTVal(formula[i])
		if err != nil {
			return nil, err
		}
		verboseStepf(stepValue, "%s  <= %q", val, formula[i])
		i++

		rv, err = ApplyOperation(rv, op, val)
		if err != nil {
			return nil, err
		}
	}

	return rv, nil
}

// ApplyOperation does a calculation using the provided arguments and returns the result.
func ApplyOperation(leftVal *DTVal, op Operation, rightVal *DTVal) (rv *DTVal, err error) {
	argsOK := false
	defer func() {
		if err != nil {
			if argsOK {
				err = fmt.Errorf("cannot apply operation %s %s %s: %w", leftVal, op, rightVal, err)
			} else {
				err = fmt.Errorf("cannot apply operation %q %q %q: %w", leftVal, op, rightVal, err)
			}
		} else {
			verboseStepf(stepResult, "%s  <= %s %s %s", rv, leftVal, op, rightVal)
		}
	}()

	if err = leftVal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid left value: %w", err)
	}
	if err = op.Validate(); err != nil {
		return nil, fmt.Errorf("invalid operation: %w", err)
	}
	if err = rightVal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid right value: %w", err)
	}
	argsOK = true

	switch op {
	case OpAdd:
		switch {
		case leftVal.IsDur() && rightVal.IsDur():
			return NewDurVal(*leftVal.Dur + *rightVal.Dur), nil
		case leftVal.IsDur() && rightVal.IsTime():
			return NewTimeVal(rightVal.Time.Add(*leftVal.Dur)), nil
		case leftVal.IsTime() && rightVal.IsDur():
			return NewTimeVal(leftVal.Time.Add(*rightVal.Dur)), nil
		case leftVal.IsNum() && rightVal.IsNum():
			return NewNumVal(*leftVal.Num + *rightVal.Num), nil
		}
	case OpSub:
		switch {
		case leftVal.IsDur() && rightVal.IsDur():
			return NewDurVal(*leftVal.Dur - *rightVal.Dur), nil
		case leftVal.IsTime() && rightVal.IsDur():
			return NewTimeVal(leftVal.Time.Add(-1 * *rightVal.Dur)), nil
		case leftVal.IsTime() && rightVal.IsTime():
			return NewDurVal(leftVal.Time.Sub(*rightVal.Time)), nil
		case leftVal.IsNum() && rightVal.IsNum():
			return NewNumVal(*leftVal.Num - *rightVal.Num), nil
		}
	case OpMul:
		switch {
		case leftVal.IsDur() && rightVal.IsNum():
			return NewDurVal(*leftVal.Dur * time.Duration(*rightVal.Num)), nil
		case leftVal.IsNum() && rightVal.IsDur():
			return NewDurVal(time.Duration(*leftVal.Num) * *rightVal.Dur), nil
		case leftVal.IsNum() && rightVal.IsNum():
			return NewNumVal(*leftVal.Num * *rightVal.Num), nil
		}
	case OpDiv:
		switch {
		case leftVal.IsDur() && rightVal.IsNum():
			return NewDurVal(*leftVal.Dur / time.Duration(*rightVal.Num)), nil
		case leftVal.IsDur() && rightVal.IsDur():
			return NewNumVal(int(*leftVal.Dur) / int(*rightVal.Dur)), nil
		case leftVal.IsNum() && rightVal.IsNum():
			return NewNumVal(*leftVal.Num / *rightVal.Num), nil
		}
	default:
		panic(fmt.Errorf("no case defined for operation %q", op))
	}

	return nil, fmt.Errorf("operation %s %s %s not defined", leftVal.TypeString(), op, rightVal.TypeString())
}
