package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const DEFAULT_COUNT = 0

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	Debugf("Parsed Input:\n%s", input)
	switch params.Option {
	case 1:
		answer, err := Explore(params, input)
		if len(answer) == 0 && err == nil {
			answer = "Stopped after exploration."
		}
		return answer, err
	case 2:
		return RunExpected(params, input)
	case 3:
		return Manual(params, input)
	case 0, 4:
		return TrySolve(params, input)
	default:
		return "", fmt.Errorf("unknown option %d", params.Option)
	}
}

func TrySolve(params *Params, input *Input) (string, error) {
	expWireMap, expGatesAll, err := CreateExpectedCircuit(input.Wires)
	if err != nil {
		return "", fmt.Errorf("could not create expected circuit: %w", err)
	}
	actWireMap, actGatesAll, err := BuildCircuit(input.Wires, input.Gates)
	if err != nil {
		return "", fmt.Errorf("could not create actual circuit: %w", err)
	}
	expWires := SortWires(expWireMap)
	actWires := SortWires(actWireMap)
	expGates := SortGates(expGatesAll)
	actGates := SortGates(actGatesAll)

	Verbosef("Validating all x and y wires.")
	if err = ValidateAllXYWires(expWires.X, expWires.Y); err != nil {
		return "", fmt.Errorf("expected: %w", err)
	}
	if err = ValidateAllXYWires(actWires.X, actWires.Y); err != nil {
		return "", fmt.Errorf("actual: %w", err)
	}
	Verbosef("Done validating all x and y wires.")

	expNameMap := make(map[string][]string)
	for i, xWire := range actWires.X {
		for j, gate := range xWire.Dests {
			expName := expWires.X[i].Dests[j].OutName
			expNameMap[expName] = append(expNameMap[expName], gate.OutName)
		}
	}
	Verbosef("Found %d name mappings from the x and y wires.", len(expNameMap))

	for _, expName := range slices.Sorted(maps.Keys(expNameMap)) {
		if len(expNameMap[expName]) != 1 {
			Stderrf("Ambiguous mappings: %q => %q", expName, expNameMap[expName])
		}
	}

	var badGates []*Gate
	for _, gate := range actGates.XOR {
		// 3 types of XOR gate:
		// X and y in, mid out to xor, and.
		// Mids in, z out.
		// x and y in, z out, only for 00.
		if gate.In1.IsXY() && gate.In2.IsXY() && gate.Out.IsMid() {
			if len(gate.Out.Dests) == 2 && gate.Out.Dests[0].Op == XOR && gate.Out.Dests[1].Op == AND {
				continue
			}
		}
		if gate.In1.IsMid() && gate.In2.IsMid() && gate.Out.IsZ() {
			continue
		}
		if gate.In1.IsXY() && gate.In2.IsXY() && gate.Out.IsZ() && gate.In1.Number == 0 {
			continue
		}
		Debugf("Found bad xor gate: %s | %c %c %c", gate, gate.In1.Type, gate.In2.Type, gate.Out.Type)
		Debugf("Gate: %#v", gate)
		Debugf("In1: %#v", gate.In1)
		Debugf("In2: %#v", gate.In2)
		Debugf("Out: %#v", gate.Out)
		badGates = append(badGates, gate)
	}
	for _, gate := range actGates.AND {
		// 2 types of AND gates:
		// X and y in, mid out to or.
		// Mids in, mid out to or.
		if len(gate.Out.Dests) == 1 && gate.Out.Dests[0].Op == OR {
			if gate.In1.IsXY() && gate.In2.IsXY() && gate.Out.IsMid() {
				continue
			}
			if gate.In1.IsMid() && gate.In2.IsMid() && gate.Out.IsMid() {
				continue
			}
		}
		Debugf("Found bad and gate: %s | %c %c %c", gate, gate.In1.Type, gate.In2.Type, gate.Out.Type)
		badGates = append(badGates, gate)
	}
	for _, gate := range actGates.OR {
		// 2 types of OR gate:
		// Mids on all, out to xor, and.
		// Mids in, z45 out.
		if gate.In1.IsMid() && gate.In2.IsMid() {
			if gate.Out.IsMid() && len(gate.Out.Dests) == 2 && gate.Out.Dests[0].Op == XOR && gate.Out.Dests[1].Op == AND {
				continue
			}
			if gate.Out.IsZ() && gate.Out.Number == 45 {
				continue
			}
		}
		Debugf("Found bad  or gate: %s | %c %c %c", gate, gate.In1.Type, gate.In2.Type, gate.Out.Type)
		badGates = append(badGates, gate)
	}

	Stderrf("Bad Gates (%d):\n%s", len(badGates), StringNumberJoin(badGates, 1, "\n"))
	_ = expGates

	rv := MapSlice(badGates, (*Gate).GetOutName)
	slices.Sort(rv)
	return strings.Join(rv, ","), nil
}

func ValidateAllXYWires(xs, ys []*Wire) error {
	if len(xs) != len(ys) {
		return fmt.Errorf("number of x wires (%d) does not equal number of y wires (%d)", len(xs), len(ys))
	}
	for i, x := range xs {
		y := ys[i]
		if err := ValidateXYWires(x, y); err != nil {
			return fmt.Errorf("[%d]: %w", i, err)
		}
	}
	return nil
}

func ValidateXYWires(x, y *Wire) error {
	Debugf("Validating %s and %s", x, y)
	if err := ValidateXYWire(x); err != nil {
		return err
	}
	if err := ValidateXYWire(y); err != nil {
		return err
	}

	for i, expOp := range []Op{XOR, AND} {
		xGate := x.Dests[i]
		yGate := y.Dests[i]
		if xGate != yGate {
			return fmt.Errorf("%s and %s have different destination gate[%d]: %s and %s", x, y, i, xGate, yGate)
		}
		if xGate.Op != expOp {
			return fmt.Errorf("%s destination gate %d is not %s: %s", x, i, expOp, xGate)
		}
		if yGate.Op != expOp {
			return fmt.Errorf("%s destination gate %d is not %s: %s", y, i, expOp, xGate)
		}
		if xGate.GetOther(x.Name).Name != y.Name {
			return fmt.Errorf("%s gate %d other input is %s, expected %s", x, i, xGate.GetOther(x.Name).Name, y.Name)
		}
		if yGate.GetOther(y.Name).Name != x.Name {
			return fmt.Errorf("%s gate %d other input is %s, expected %s", y, i, yGate.GetOther(y.Name).Name, x.Name)
		}
	}
	return nil
}

func ValidateXYWire(wire *Wire) error {
	if wire.Source != nil {
		return fmt.Errorf("%s has a source: %s", wire, wire.Source)
	}

	if len(wire.Dests) != 2 {
		return fmt.Errorf("%s has %d destinations", wire, len(wire.Dests))
	}
	if wire.Dests[0].Op != XOR {
		return fmt.Errorf("%s first destination is not XOR", wire)
	}
	if wire.Dests[1].Op != AND {
		return fmt.Errorf("%s second destination is not AND", wire)
	}

	return nil
}

type SortedGates struct {
	XOR []*Gate
	AND []*Gate
	OR  []*Gate
}

func SortGates(gates []*Gate) *SortedGates {
	rv := &SortedGates{}
	for _, gate := range gates {
		switch gate.Op {
		case XOR:
			rv.XOR = append(rv.XOR, gate)
		case AND:
			rv.AND = append(rv.AND, gate)
		case OR:
			rv.OR = append(rv.OR, gate)
		default:
			panic(fmt.Errorf("gate %s has unknown op %q", gate, gate.Op))
		}
	}
	return rv
}

func Manual(params *Params, input *Input) (string, error) {
	var swaps map[string]string
	var wireNamesOfInterest []string
	if len(params.Custom) > 0 {
		var err error
		swaps, wireNamesOfInterest, err = ParseCustom(params.Custom)
		if err != nil {
			return "", err
		}
	}
	swapStrs := make([]string, 0, len(swaps))
	for w1Name, w2Name := range swaps {
		swapStrs = append(swapStrs, w1Name+"-"+w2Name)
	}

	wires := make([]*Wire, len(input.Wires))
	for i, wire := range input.Wires {
		wires[i] = wire.FreshCopy()
		switch wire.Type {
		case X:
			wires[i].Value = "1"
		case Y:
			wires[i].Value = "0"
		}
	}

	wireMap, gates, err := BuildCircuit(wires, input.Gates)
	if err != nil {
		return "", fmt.Errorf("could not build circuit: %w", err)
	}
	result, err := RunCircuit(wireMap, gates, swaps)
	if err != nil {
		return "", err
	}

	if !verbose {
		// This is printed in RunCircuit if verbose is on, but we want it no matter what.
		Stderrf("Result of swaps: %s\n%s", swapStrs, result)
	}

	for i := len(result.ExpZBin) - 1; i >= 0; i-- {
		if result.ExpZBin[i] != result.ZBin[i] {
			wireNamesOfInterest = append(wireNamesOfInterest, fmt.Sprintf("z%02d", i))
			break
		}
	}

	for _, name := range wireNamesOfInterest {
		wire := result.WireMap[name]
		if wire == nil {
			Stderrf("%s: no such wire.", name)
			continue
		}
		Stderrf("%s>: %s", name, TraceWireForward(wire))
		Stderrf("%s<: %s", name, TraceWireBackward(wire))
		if debug {
			Stderrf("%s<+: %s", name, TraceEqBackward(wire))
		}
	}

	return "", nil
}

func RunExpected(params *Params, input *Input) (string, error) {
	var swaps map[string]string
	var wireNamesOfInterest []string
	if len(params.Custom) > 0 {
		var err error
		swaps, wireNamesOfInterest, err = ParseCustom(params.Custom)
		if err != nil {
			return "", err
		}
	}

	wireMap, gates, err := CreateExpectedCircuit(input.Wires)
	if err != nil {
		return "", err
	}

	result, err := RunCircuit(wireMap, gates, swaps)
	if err != nil {
		return "", err
	}

	for _, name := range wireNamesOfInterest {
		wire := result.WireMap[name]
		if wire == nil {
			Stderrf("%s: no such wire.", name)
			continue
		}
		Stderrf("%s>: %s", name, TraceWireForward(wire))
		Stderrf("%s<: %s", name, TraceWireBackward(wire))
		if debug {
			Stderrf("%s<+: %s", name, TraceEqBackward(wire))
		}
	}

	return "Done running expected.", nil
}

func Explore(params *Params, input *Input) (string, error) {
	var swaps map[string]string
	var wireNamesOfInterest []string
	if len(params.Custom) > 0 {
		var err error
		swaps, wireNamesOfInterest, err = ParseCustom(params.Custom)
		if err != nil {
			return "", err
		}
	}

	wireMap, gates, err := BuildCircuit(input.Wires, input.Gates)
	if err != nil {
		return "", fmt.Errorf("could not build circuit: %w", err)
	}
	result, err := RunCircuit(wireMap, gates, swaps)
	if err != nil {
		return "", err
	}

	for _, name := range wireNamesOfInterest {
		wire := result.WireMap[name]
		if wire == nil {
			Stderrf("%s: no such wire.", name)
			continue
		}
		Stderrf("%s>: %s", name, TraceWireForward(wire))
		Stderrf("%s<: %s", name, TraceWireBackward(wire))
		if debug {
			Stderrf("%s<+: %s", name, TraceEqBackward(wire))
		}
	}

	return "", nil
}

func CreateExpectedGateStrings() []string {
	rv := make([]string, 2, 222)
	rv[0] = "x00 XOR y00 => z00"
	rv[1] = "x00 AND y00 => C00"
	for i := 1; i <= 44; i++ {
		x := fmt.Sprintf("x%02d", i)
		y := fmt.Sprintf("y%02d", i)
		z := fmt.Sprintf("z%02d", i)
		p := fmt.Sprintf("P%02d", i)
		q := fmt.Sprintf("Q%02d", i)
		r := fmt.Sprintf("R%02d", i)
		c0 := fmt.Sprintf("C%02d", i-1)
		c1 := fmt.Sprintf("C%02d", i)
		if i == 44 {
			c1 = fmt.Sprintf("z%02d", i+1)
		}
		rv = append(rv,
			fmt.Sprintf("%s %s %s => %s", x, XOR, y, p),
			fmt.Sprintf("%s %s %s => %s", x, AND, y, q),
			fmt.Sprintf("%s %s %s => %s", p, XOR, c0, z),
			fmt.Sprintf("%s %s %s => %s", p, AND, c0, r),
			fmt.Sprintf("%s %s %s => %s", q, OR, r, c1),
		)
	}
	return rv
}

func CreateExpectedCircuit(wires []*Wire) (map[string]*Wire, []*Gate, error) {
	gateStrs := CreateExpectedGateStrings()

	gates := make([]*Gate, len(gateStrs))
	var err error
	for i, gateStr := range gateStrs {
		gates[i], err = ParseGate(gateStr)
		if err != nil {
			return nil, nil, err
		}
	}

	return BuildCircuit(wires, gates)
}

func TraceEqBackward(wire *Wire) string {
	if wire.Source == nil {
		return wire.Name
	}
	in1 := TraceEqBackward(wire.Source.In1)
	if strings.Contains(in1, " ") {
		in1 = "(" + in1 + ")"
	}
	in2 := TraceEqBackward(wire.Source.In2)
	if strings.Contains(in2, " ") {
		in2 = "(" + in2 + ")"
	}
	return fmt.Sprintf("%s %s %s => %s", in1, wire.Source.Op, in2, wire.Name)
}

type WireTrace []*Wire

func (t WireTrace) String() string {
	counts := make(map[string]int)
	var order []string
	for _, wire := range t {
		if counts[wire.Name] == 0 {
			order = append(order, wire.Name)
		}
		counts[wire.Name]++
	}
	parts := make([]string, len(order))
	for i, name := range order {
		parts[i] = fmt.Sprintf("%s*%d", name, counts[name])
	}
	return fmt.Sprintf("(%d): %s", len(t), strings.Join(parts, " "))
}

func TraceWireBackward(wire *Wire) WireTrace {
	rv := WireTrace{wire}
	if wire.Source == nil {
		return rv
	}
	in1 := TraceWireBackward(wire.Source.In1)
	in2 := TraceWireBackward(wire.Source.In2)
	rv = append(rv, in1...)
	rv = append(rv, in2...)
	slices.SortFunc(rv, CompareWires)
	return rv
}

func TraceWireForward(wire *Wire) WireTrace {
	rv := WireTrace{wire}
	for _, dep := range wire.Dests {
		rv = append(rv, TraceWireForward(dep.Out)...)
	}
	slices.SortFunc(rv, CompareWires)
	return rv
}

func ParseCustom(lines []string) (map[string]string, []string, error) {
	if len(lines) == 0 {
		return nil, nil, nil
	}
	swaps := make(map[string]string)
	var wires []string
	for i, line := range lines {
		fields := strings.Fields(line)
		for f, field := range fields {
			parts := strings.Split(field, "-")
			switch len(parts) {
			case 1:
				wires = append(wires, parts[0])
			case 2:
				swaps[parts[0]] = parts[1]
			default:
				return nil, nil, fmt.Errorf("could not parse custom line %d %q, field %d %q: unknown format", i, line, f, field)
			}
		}
	}
	return swaps, wires, nil
}

type Result struct {
	X      int64
	XBin   string
	XWires []*Wire

	Y      int64
	YBin   string
	YWires []*Wire

	ExpZ    int64
	ExpZBin string

	Z      int64
	ZBin   string
	ZWires []*Wire

	WireMap map[string]*Wire
	Gates   []*Gate
}

func (s *Result) String() string {
	blen := MaxLen(s.XBin, s.YBin, s.ExpZBin, s.ZBin)
	dfmt := DigitFormatForMax(int(max(s.X, max(s.Y, max(s.ExpZ, s.Z)))))
	lines := []string{
		fmt.Sprintf("  x: "+dfmt+" = %s", s.X, PadLeft(s.XBin, blen)),
		fmt.Sprintf("  y: "+dfmt+" = %s", s.Y, PadLeft(s.YBin, blen)),
		fmt.Sprintf("x+y: "+dfmt+" = %s", s.ExpZ, PadLeft(s.ExpZBin, blen)),
		fmt.Sprintf("  z: "+dfmt+" = %s", s.Z, PadLeft(s.ZBin, blen)),
	}
	rulerLabel := "ruler: "
	ruler1 := "98765432109876543210987654321098765432109876543210"
	ruler2 := "4444444444333333333322222222221111111111          "
	ruler1 = ruler1[len(ruler1)-blen:]
	space1 := strings.Repeat(" ", len(lines[3])-len(ruler1)-len(rulerLabel))
	lines = append(lines, space1+rulerLabel+ruler1)
	if blen > 10 {
		ruler2 = ruler2[len(ruler2)-blen:]
		space2 := strings.Repeat(" ", len(lines[3])-len(ruler2))
		lines = append(lines, space2+ruler2)
	}
	return strings.Join(lines, "\n")
}

type SortedWires struct {
	X   []*Wire
	Y   []*Wire
	Z   []*Wire
	Mid []*Wire
}

func SortWires(wireMap map[string]*Wire) *SortedWires {
	rv := &SortedWires{}
	for name, wire := range wireMap {
		switch name[0] {
		case 'x':
			rv.X = append(rv.X, wire)
		case 'y':
			rv.Y = append(rv.Y, wire)
		case 'z':
			rv.Z = append(rv.Z, wire)
		default:
			rv.Mid = append(rv.Mid, wire)
		}
	}
	slices.SortFunc(rv.X, CompareWires)
	slices.SortFunc(rv.Y, CompareWires)
	slices.SortFunc(rv.Z, CompareWires)
	slices.SortFunc(rv.Mid, CompareWires)
	return rv
}

func RunCircuit(wireMap map[string]*Wire, gates []*Gate, swaps map[string]string) (*Result, error) {
	sortedWires := SortWires(wireMap)
	xWires := sortedWires.X
	yWires := sortedWires.Y
	zWires := sortedWires.Z
	midWires := sortedWires.Mid

	swapStrs := make([]string, 0, len(swaps))
	for w1Name, w2Name := range swaps {
		swapStrs = append(swapStrs, w1Name+"-"+w2Name)
		wire1 := wireMap[w1Name]
		if wire1 == nil {
			return nil, fmt.Errorf("wire1 %q does not exist to swap", w1Name)
		}
		wire2 := wireMap[w2Name]
		if wire2 == nil {
			return nil, fmt.Errorf("wire2 %q does not exist to swap", w2Name)
		}
		SwapWires(wire1, wire2)
	}

	PropagateSignals(gates, zWires)
	if debug {
		Stderrf("x wires (%d):\n%s", len(xWires), StringNumberJoin(xWires, 1, "\n"))
		Stderrf("y wires (%d):\n%s", len(yWires), StringNumberJoin(yWires, 1, "\n"))
		Stderrf("z wires (%d):\n%s", len(zWires), StringNumberJoin(zWires, 1, "\n"))
		Stderrf("mid wires (%d):\n%s", len(midWires), StringNumberJoin(midWires, 1, "\n"))
	}

	xBin := strings.Join(MapSlice(xWires, (*Wire).GetValue), "")
	yBin := strings.Join(MapSlice(yWires, (*Wire).GetValue), "")
	zBin := strings.Join(MapSlice(zWires, (*Wire).GetValue), "")
	x, err := strconv.ParseInt(xBin, 2, 64)
	if err != nil {
		return nil, fmt.Errorf("could not convert x values %q to int: %w", xBin, err)
	}
	y, err := strconv.ParseInt(yBin, 2, 64)
	if err != nil {
		return nil, fmt.Errorf("could not convert y values %q to int: %w", yBin, err)
	}
	z, err := strconv.ParseInt(zBin, 2, 64)
	if err != nil {
		return nil, fmt.Errorf("could not convert z values %q to int: %w", zBin, err)
	}
	expZ := x + y
	expZBin := strconv.FormatInt(expZ, 2)
	if len(expZBin) < len(zBin) {
		expZBin = "0" + expZBin
	}

	rv := &Result{
		X:       x,
		XBin:    xBin,
		XWires:  xWires,
		Y:       y,
		YBin:    yBin,
		YWires:  yWires,
		ExpZ:    expZ,
		ExpZBin: expZBin,
		Z:       z,
		ZBin:    zBin,
		ZWires:  zWires,
		WireMap: wireMap,
		Gates:   gates,
	}
	Verbosef("result of swaps: %s\n%s", swapStrs, rv)

	return rv, nil
}

func SwapOuts(gate1, gate2 *Gate) {
	gate1.Out, gate2.Out = gate2.Out, gate1.Out
	gate1.Out.Source = gate1
	gate2.Out.Source = gate2
}

func SwapWires(wire1, wire2 *Wire) {
	wire1.Source, wire2.Source = wire2.Source, wire1.Source
	wire1.Source.Out = wire1
	wire2.Source.Out = wire2
}

func PropagateSignals(gates []*Gate, keyWires []*Wire) {
	for !AllHaveValue(keyWires) {
		redoGates := make([]*Gate, 0, len(gates))
		for _, gate := range gates {
			if !gate.Propagate() {
				redoGates = append(redoGates, gate)
			}
		}
		if len(gates) == len(redoGates) {
			panic(errors.New("could not propagate any signals"))
		}
		gates = redoGates
	}
}

func AllHaveValue(wires []*Wire) bool {
	for _, wire := range wires {
		if len(wire.Value) == 0 {
			return false
		}
	}
	return true
}

func BuildCircuit(wires []*Wire, gates []*Gate) (map[string]*Wire, []*Gate, error) {
	wires = MapSlice(wires, (*Wire).FreshCopy)
	gates = MapSlice(gates, (*Gate).FreshCopy)
	wireMap := make(map[string]*Wire)
	for _, wire := range wires {
		wireMap[wire.Name] = wire
	}

	for _, gate := range gates {
		if wireMap[gate.In1Name] == nil {
			wireMap[gate.In1Name] = NewWire(gate.In1Name)
		}
		if wireMap[gate.In2Name] == nil {
			wireMap[gate.In2Name] = NewWire(gate.In2Name)
		}
		if wireMap[gate.OutName] == nil {
			wireMap[gate.OutName] = NewWire(gate.OutName)
		}

		gate.In1 = wireMap[gate.In1Name]
		gate.In2 = wireMap[gate.In2Name]
		gate.Out = wireMap[gate.OutName]
		if gate.Out.Source != nil {
			return nil, nil, fmt.Errorf("could not wire up %s: wire out already has a destination", gate)
		}

		gate.In1.Dests = append(gate.In1.Dests, gate)
		gate.In2.Dests = append(gate.In2.Dests, gate)
		gate.Out.Source = gate
	}

	for _, wire := range wires {
		SortDests(wire)
	}

	return wireMap, gates, nil
}

func SortDests(wire *Wire) {
	slices.SortFunc(wire.Dests, func(a, b *Gate) int {
		if a == b {
			return 0
		}
		if a == nil {
			return 1
		}
		if b == nil {
			return -1
		}
		if rv := CompareOps(a.Op, b.Op); rv != 0 {
			return rv
		}
		if rv := CompareWires(a.GetOther(wire.Name), b.GetOther(wire.Name)); rv != 0 {
			return rv
		}
		return CompareWires(a.Out, b.Out)
	})
}

type WireType byte

const (
	X   = WireType('x')
	Y   = WireType('y')
	Z   = WireType('z')
	Mid = WireType('~')
)

var WireTypeOrder = map[WireType]int{X: 1, Y: 2, Mid: 3, Z: 4}

func CompareWireTypes(a, b WireType) int {
	if a == b {
		return 0
	}
	oa, oka := WireTypeOrder[a]
	ob, okb := WireTypeOrder[b]
	switch {
	case oka && okb:
		return CompareInts(oa, ob)
	case oka:
		return -1
	case okb:
		return 1
	}
	return 0
}

func CompareInts[N Integer](a, b N) int {
	if a == b {
		return 0
	}
	if a < b {
		return -1
	}
	return 1
}

type Wire struct {
	Name   string
	Value  string
	Source *Gate
	Dests  []*Gate
	Type   WireType
	Number int
}

func NewWire(name string) *Wire {
	rv := &Wire{Name: name}
	switch name[0] {
	case 'x', 'y', 'z':
		rv.Type = WireType(name[0])
		var err error
		rv.Number, err = strconv.Atoi(name[1:])
		if err != nil {
			panic(fmt.Errorf("could not extract number from %q: %w", name, err))
		}
	default:
		rv.Type = Mid
	}
	return rv
}

func (w *Wire) WithValue(value string) *Wire {
	w.Value = value
	return w
}

func (w *Wire) String() string {
	dests := make([]string, len(w.Dests))
	for i, gate := range w.Dests {
		other := Ternary(gate.In1Name != w.Name, gate.In1Name, gate.In2Name)
		dests[i] = fmt.Sprintf("%3s %s => %s", gate.Op, other, gate.OutName)
	}
	destStr := ""
	if len(dests) > 0 {
		destStr = " => " + strings.Join(dests, " | ")
	}
	srcStr := w.Value
	if w.Source != nil {
		srcStr = w.Source.String()
	}
	return fmt.Sprintf("%s(%d)=%s: %s%s", w.Name, w.Number, w.Value, srcStr, destStr)
}

func (w *Wire) GetValue() string {
	return w.Value
}

func (w *Wire) GetName() string {
	return w.Name
}

func (w *Wire) FreshCopy() *Wire {
	if w == nil {
		return nil
	}
	return NewWire(w.Name).WithValue(w.Value)
}

func (w *Wire) IsXY() bool {
	return w.Type == X || w.Type == Y
}

func (w *Wire) IsZ() bool {
	return w.Type == Z
}

func (w *Wire) IsMid() bool {
	return w.Type == Mid
}

func ParseWire(line string) (*Wire, error) {
	parts := strings.Split(line, ": ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid wire line %q: expected format '<name>: <value>'", line)
	}
	if parts[1] != "0" && parts[1] != "1" {
		return nil, fmt.Errorf("invalid wire line %q: value %q is illegal", line, parts[1])
	}
	return NewWire(parts[0]).WithValue(parts[1]), nil
}

func CompareWires(a, b *Wire) int {
	if a == b {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}
	// Sort by type first, x, y, mid, z.
	if rv := CompareWireTypes(a.Type, b.Type); rv != 0 {
		return rv
	}
	// Now sort by reverse number for x, y, and z types, or normal name for mids.
	if a.Type == Mid {
		return strings.Compare(a.Name, b.Name)
	}
	return CompareInts(a.Number, b.Number) * -1
}

type Op string

const (
	AND = Op("AND")
	OR  = Op("OR")
	XOR = Op("XOR")

	Zero = "0"
	One  = "1"
)

func IsOp(val string) bool {
	op := Op(val)
	if op == AND || op == OR || op == XOR {
		return true
	}
	return false
}

func CompareOps(a, b Op) int {
	if a == b {
		return 0
	}
	if a == XOR {
		return -1
	}
	if b == XOR {
		return 1
	}
	if a == AND {
		return -1
	}
	if b == AND {
		return 1
	}
	if a == OR {
		return -1
	}
	if b == OR {
		return 1
	}
	return strings.Compare(string(a), string(b))
}

type Gate struct {
	Name    string
	In1Name string
	Op      Op
	In2Name string
	OutName string
	In1     *Wire
	In2     *Wire
	Out     *Wire
	Value   string
}

func (g *Gate) String() string {
	return fmt.Sprintf("%s %3s %s %q => %s", g.In1Name, g.Op, g.In2Name, g.Value, g.OutName)
}

func (g *Gate) FreshCopy() *Gate {
	return &Gate{
		Name:    g.Name,
		In1Name: g.In1Name,
		Op:      g.Op,
		In2Name: g.In2Name,
		OutName: g.OutName,
	}
}

func (g *Gate) Propagate() bool {
	if len(g.Value) != 0 || len(g.In1.Value) == 0 || len(g.In2.Value) == 0 {
		return false
	}

	switch g.Op {
	case AND:
		g.Value = Ternary(g.In1.Value == One && g.In2.Value == One, One, Zero)
	case OR:
		g.Value = Ternary(g.In1.Value == One || g.In2.Value == One, One, Zero)
	case XOR:
		g.Value = Ternary(g.In1.Value != g.In2.Value, One, Zero)
	default:
		panic(fmt.Errorf("gate has unknown operation %q", g.Op))
	}

	g.Out.Value = g.Value
	return true
}

func (g *Gate) GetOther(name string) *Wire {
	if g.In1Name == name {
		return g.In2
	}
	if g.In2Name == name {
		return g.In1
	}
	return nil
}

func (g *Gate) GetOutName() string {
	return g.OutName
}

func ParseGate(line string) (*Gate, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 5 {
		return nil, fmt.Errorf("invalid gate line %q: expected format '<wire> <op> <wire> -> <wire>'", line)
	}
	if !IsOp(parts[1]) {
		return nil, fmt.Errorf("invalid gate line %q: unknown operation %q", line, parts[1])
	}
	if len(parts[0]) == 0 || len(parts[2]) == 0 || len(parts[4]) == 0 {
		return nil, fmt.Errorf("invalid gate line %q: empty wire", line)
	}
	return &Gate{Name: line, In1Name: parts[0], Op: Op(parts[1]), In2Name: parts[2], OutName: parts[4]}, nil
}

type Input struct {
	Wires []*Wire
	Gates []*Gate
}

func (i Input) String() string {
	// StringNumberJoin(slice, startAt, sep) string
	// StringNumberJoinFunc(slice, stringer, startAt, sep) string
	// SliceToStrings(slice) []string
	// AddLineNumbers(lines, startAt) []string
	// MapSlice(slice, mapper) slice  or  MapPSlice  or  MapSliceP
	// CreateIndexedGridString(grid, color, highlight) string  or  CreateIndexedGridStringBz  or  CreateIndexedGridStringNums
	// CreateIndexedGridStringFunc(grid, converter, color, highlight)
	return fmt.Sprintf("Wires (%d):\n%s\n\nGates (%d):\n%s",
		len(i.Wires), StringNumberJoin(i.Wires, 1, "\n"),
		len(i.Gates), StringNumberJoin(i.Gates, 1, "\n"))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{}
	inGates := false
	for i, line := range lines {
		if len(line) == 0 {
			inGates = true
			continue
		}
		if !inGates {
			wire, err := ParseWire(line)
			if err != nil {
				return &rv, fmt.Errorf("[%d]: %w", i, err)
			}
			rv.Wires = append(rv.Wires, wire)
			continue
		}
		gate, err := ParseGate(line)
		if err != nil {
			return &rv, fmt.Errorf("[%d]: %w", i, err)
		}
		rv.Gates = append(rv.Gates, gate)
	}
	return &rv, nil
}

func MaxLen(strs ...string) int {
	rv := 0
	for _, str := range strs {
		rv = max(rv, len(str))
	}
	return rv
}

// PadLeft will return a string with spaces added to the left of the provided one up to the provided length.
func PadLeft(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return strings.Repeat(" ", length-len(str)) + str
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------  Some generic stuff  --------------------------------------
// -------------------------------------------------------------------------------------------------

const (
	MIN_INT8  = int8(-128)
	MAX_INT8  = int8(127)
	MIN_INT16 = int16(-32_768)
	MAX_INT16 = int16(32_767)
	MIN_INT32 = int32(-2_147_483_648)
	MAX_INT32 = int32(2_147_483_647)
	MIN_INT64 = int64(-9_223_372_036_854_775_808)
	MAX_INT64 = int64(9_223_372_036_854_775_807)
	MIN_INT   = -9_223_372_036_854_775_808
	MAX_INT   = 9_223_372_036_854_775_807

	MAX_UINT8  = uint8(255)
	MAX_UINT16 = uint16(65_535)
	MAX_UINT32 = uint32(4_294_967_295)
	MAX_UINT64 = uint64(18_446_744_073_709_551_615)
	MAX_UINT   = uint(18_446_744_073_709_551_615)

	NilStr = "<nil>"
)

// SplitParseInts splits a string on whitespace and converts each part into an int.
// Uses strings.Fields(s) for the splitting and strconv.Atoi to parse it to an int.
// Leading and trailing whitespace on each entry are ignored.
func SplitParseInts(s string) ([]int, error) {
	rv := []int{}
	for _, entry := range strings.Fields(s) {
		if len(entry) > 0 {
			i, err := strconv.Atoi(strings.TrimSpace(entry))
			if err != nil {
				return rv, err
			}
			rv = append(rv, i)
		}
	}
	return rv, nil
}

// SplitParseIntsD splits a string on the provided delimiter and converts each part into an int.
// Uses strings.Fields(s) for the splitting and strconv.Atoi to parse it to an int.
// Leading and trailing whitespace on each entry are ignored.
func SplitParseIntsD(s, d string) ([]int, error) {
	rv := []int{}
	for _, entry := range strings.Split(s, d) {
		if len(entry) > 0 {
			i, err := strconv.Atoi(strings.TrimSpace(entry))
			if err != nil {
				return rv, err
			}
			rv = append(rv, i)
		}
	}
	return rv, nil
}

// StringJoin maps the slice to strings and joins them.
func StringJoin[S ~[]E, E fmt.Stringer](slice S, sep string) string {
	return strings.Join(MapSlice(slice, E.String), sep)
}

// StringNumberJoin maps the slice to strings, numbers them, and joins them.
func StringNumberJoin[S ~[]E, E fmt.Stringer](slice S, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, E.String), startAt), sep)
}

// StringNumberJoinFunc maps the slice to strings using the provided stringer, numbers them, and joins them.
func StringNumberJoinFunc[S ~[]E, E any](slice S, stringer func(E) string, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, stringer), startAt), sep)
}

// SliceToStrings runs String() on each entry of the provided slice.
func SliceToStrings[S ~[]E, E fmt.Stringer](slice S) []string {
	return MapSlice(slice, E.String)
}

// AddLineNumbers adds line numbers to each string.
func AddLineNumbers(lines []string, startAt int) []string {
	if len(lines) == 0 {
		return []string{}
	}
	lineFmt := DigitFormatForMax(len(lines)-1+startAt) + ": %s"
	rv := make([]string, len(lines))
	for i, line := range lines {
		rv[i] = fmt.Sprintf(lineFmt, i+startAt, line)
	}
	return rv
}

// DigitFormatForMax returns a format string of the length of the provided maximum number.
// E.g. DigitFormatForMax(10) returns "%2d".
// DigitFormatForMax(382920) returns "%6d".
func DigitFormatForMax(maximum int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", maximum)))
}

// PrefixLines splits each provided string on \n then adds a prefix to each line, then puts it all back together.
func PrefixLines(pre string, strs ...string) string {
	var rv strings.Builder
	lastI := len(strs) - 1
	for i, str := range strs {
		lines := strings.Split(str, "\n")
		lastJ := len(lines) - 1
		for j, line := range lines {
			rv.WriteString(pre)
			rv.WriteString(line)
			if i != lastI || j != lastJ {
				rv.WriteByte('\n')
			}
		}
	}
	return rv.String()
}

// MapSlice returns a new slice with each element run through the provided mapper function.
// Use MapSlice if the slice and mapper are either both concrete or both pointers.
// Use MapPSlice if the slice is pointers, but the mapper takes in a concrete E.
// Use MapSliceP if the slice is concrete, but the mapper takes in a pointer to E.
func MapSlice[S ~[]E, E any, R any](slice S, mapper func(E) R) []R {
	if slice == nil {
		return nil
	}
	rv := make([]R, len(slice))
	for i, e := range slice {
		rv[i] = mapper(e)
	}
	return rv
}

// MapPSlice returns a new slice with each element run through the provided mapper function.
// Use MapSlice if the slice and mapper are either both concrete or both pointers.
// Use MapPSlice if the slice is pointers, but the mapper takes in a concrete E.
// Use MapSliceP if the slice is concrete, but the mapper takes in a pointer to E.
func MapPSlice[S ~[]*E, E any, R any](slice S, mapper func(E) R) []R {
	if slice == nil {
		return nil
	}
	rv := make([]R, len(slice))
	for i, e := range slice {
		rv[i] = mapper(*e)
	}
	return rv
}

// MapSliceP returns a new slice with each element run through the provided mapper function.
// Use MapSlice if the slice and mapper are either both concrete or both pointers.
// Use MapPSlice if the slice is pointers, but the mapper takes in a concrete E.
// Use MapSliceP if the slice is concrete, but the mapper takes in a pointer to E.
func MapSliceP[S ~[]E, E any, R any](slice S, mapper func(*E) R) []R {
	if slice == nil {
		return nil
	}
	rv := make([]R, len(slice))
	for i, e := range slice {
		e := e
		rv[i] = mapper(&e)
	}
	return rv
}

// MakeZeroGrid creates a 2-d matrix with the given width and height, each entry having the zero value for the type.
// Usage: grid := MakeZeroGrid[byte](10, 10)
func MakeZeroGrid[V any](width, height int) [][]V {
	rv := make([][]V, height)
	for y := range rv {
		rv[y] = make([]V, width)
	}
	return rv
}

// MakeGrid creates a new 2-d matrix with the given width and height, each entry having the provided value.
func MakeGrid[V any](width, height int, value V) [][]V {
	rv := make([][]V, height)
	for y := range rv {
		rv[y] = make([]V, width)
		for x := range rv[y] {
			rv[y][x] = value
		}
	}
	return rv
}

// MapGrid creates a new map by running the provided mapper on each element of the provided grid.
func MapGrid[G ~[][]E, E any, R any](grid G, mapper func(E) R) [][]R {
	if grid == nil {
		return nil
	}
	rv := make([][]R, len(grid))
	for y := range grid {
		rv[y] = make([]R, len(grid[y]))
		for x := range rv[y] {
			rv[y][x] = mapper(grid[y][x])
		}
	}
	return rv
}

// Abs returns the absolute value of the provided number.
func Abs[V Number](v V) V {
	var zero V
	if v < zero {
		return zero - v
	}
	return v
}

// Ternary returns ifTrue if test == true, otherwise, returns ifFalse.
func Ternary[E any](test bool, ifTrue, ifFalse E) E {
	if test {
		return ifTrue
	}
	return ifFalse
}

// CopyAppend returns a copy of s with the other provided entries appended.
func CopyAppend[S ~[]E, E any](s S, es ...E) S {
	rv := make(S, len(s)+len(es))
	copy(rv, s)
	copy(rv[len(s):], es)
	return rv
}

// Alternates: ©®¬ÆæØøÞþ

// ConversionRunes are some chars used to represent numbers for smaller output. See also: GetRune.
var ConversionRunes = []rune("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz~-_+=|¦:;!@#$¢£¥%^&*()[]{}<>«»/?¿÷°§¶¤")

// GetRune returns the rune used to represent the provided number for smaller output.
// The runes will repeat every 100. E.g. IntRune(3) returns the same as IntRune(103).
func GetRune(i int) rune {
	return ConversionRunes[i%len(ConversionRunes)]
}

// Signed is a constraint of signed integer types. Same as golang.org/x/exp/constraints.Signed.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Signed is a constraint of unsigned integer types. Same as golang.org/x/exp/constraints.Unsigned.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is a constraint of integer types. Same as golang.org/x/exp/constraints.Integer.
type Integer interface {
	Signed | Unsigned
}

// Float is a constraint of float types. Same as golang.org/x/exp/constraints.Float.
type Float interface {
	~float32 | ~float64
}

// Ordered is a constraint for types that can be compared using > etc. Same as golang.org/x/exp/constraints.Ordered.
type Ordered interface {
	Integer | Float | ~string
}

// Complex is a constraint of complex types. Same as golang.org/x/exp/constraints.Complex.
type Complex interface {
	~complex64 | ~complex128
}

// Number is a constraint of integers and floats.
type Number interface {
	Integer | Float
}

// -----------------------------------------------------------------------------
// -------------------------------  Coordinates  -------------------------------
// -----------------------------------------------------------------------------

// XY is something that has an X and Y value.
type XY interface {
	GetX() int
	GetY() int
	GetXY() (int, int)
}

// A Point contains an X and Y value.
type Point struct {
	X int
	Y int
}

// NewPoint creates a new Point with the given coordinates.
func NewPoint(x, y int) *Point {
	return &Point{X: x, Y: y}
}

// String returns a string of this point in the format "(x,y)".
func (p Point) String() string {
	return fmt.Sprintf("(%d,%d)", p.X, p.Y)
}

// GetX gets this Point's X value.
func (p Point) GetX() int {
	return p.X
}

// GetY gets this Point's Y value.
func (p Point) GetY() int {
	return p.Y
}

// GetXY gets this Point's (X, Y) values.
func (p Point) GetXY() (int, int) {
	return p.X, p.Y
}

// AddPoints returns a new point that is the sum of the provided points.
// See also: AddXYs, SumXYs.
func AddPoints(points ...*Point) *Point {
	rv := NewPoint(0, 0)
	for _, p := range points {
		rv.X += p.X
		rv.Y += p.Y
	}
	return rv
}

// AddXYs returns a new point that is the sum of the provided points.
// See also: AddPoints, SumXYs.
func AddXYs(points ...XY) *Point {
	return SumXYs(points)
}

// SumXYs returns a new point that is the sum of the provided points.
// See also: AddXYs, AddPoints.
func SumXYs[S ~[]E, E XY](points S) *Point {
	rv := NewPoint(0, 0)
	for _, p := range points {
		rv.X += p.GetX()
		rv.Y += p.GetY()
	}
	return rv
}

// IsSameXY returns true if a and b have the same x and y.
func IsSameXY(a, b XY) bool {
	return a.GetX() == b.GetX() && a.GetY() == b.GetY()
}

// AsPoints converts a slice of something with an X and Y into a slice of points.
func AsPoints[S ~[]E, E XY](vals S) []*Point {
	rv := make([]*Point, len(vals))
	for i, val := range vals {
		rv[i] = NewPoint(val.GetX(), val.GetY())
	}
	return rv
}

// HasPoints returns true if there's a point in path with the same (x,y) as the point provided.
func HasPoint[S ~[]E, E XY](path S, point XY) bool {
	x, y := point.GetXY()
	for _, p := range path {
		if x == p.GetX() && y == p.GetY() {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// ------------------------------  String Makers  ------------------------------
// -----------------------------------------------------------------------------

// CreateIndexedGridStringBz is for [][]byte
// CreateIndexedGridStringNums is for [][]int or [][]uint or [][]int16 etc.
// CreateIndexedGridString is for [][]string
// All of them have the signature (vals, color, highlight)
// CreateIndexedGridStringFunc is for any other [][]; signature = (vals, converter, color, highlight)
// PointsString is simpler than PathString.
// EnhancedPathString is a similar signature as CreateIndexedGridString.

// PointString returns the "(%d,%d)" string for the provided XY.
// The generic here might seem silly, but without it, its use in PointsString gives a syntax error.
func PointString[V XY](p V) string {
	return fmt.Sprintf("(%d,%d)", p.GetX(), p.GetY())
}

// PointsString creates a string of the provided points, e.g. "(0,0) (0,1) (1,1)".
func PointsString[S ~[]E, E XY](points S) string {
	return strings.Join(MapSlice(points, PointString), " ")
}

// PathString returns a one-line string containing all the points in the provided slice of XY. E.g. "{3}[0:(0,0);1:(0,1);2:(1,1)]".
func PathString[S ~[]E, E XY](path S) string {
	lines := MapSlice(path, PointString)
	for i, line := range lines {
		lines[i] = strconv.Itoa(i) + ":" + line
	}
	return fmt.Sprintf("{%d}[%s]", len(path), strings.Join(lines, ";"))
}

// EnhancedPathString creates a string of the provided path, coloring and/or highlighting points as provided.
func EnhancedPathString[P ~[]Q, Q XY, S ~[]E, E XY](path P, colorPoints, highlightPoints S) string {
	fmts := make([]int, len(path))
	for i, point := range path {
		if HasPoint(colorPoints, point) {
			fmts[i] = 1
		}
		if HasPoint(highlightPoints, point) {
			fmts[i] += 2
		}
	}

	var rv strings.Builder
	for i, point := range path {
		if i != 0 {
			rv.WriteByte(';')
		}
		cell := fmt.Sprintf("%d:(%d,%d)", i, point.GetX(), point.GetY())
		switch fmts[i] {
		case 0: // default look.
			rv.WriteString(cell)
		case 1: // color only.
			rv.WriteString("\033[94m" + cell + "\033[0m") // Light-blue text.
		case 2: // highlight only
			rv.WriteString("\033[7m" + cell + "\033[0m") // Foreground<->Background Reversed.
		case 3: // color and highlight
			rv.WriteString("\033[94;7m" + cell + "\033[0m") // Light-blue background after fg<->bg reversed.
		default: // Unknown, make it ugly.
			rv.WriteString("\033[93;41m" + cell + "\033[0m") // Bright yellow text on a red background.
		}
	}

	return fmt.Sprintf("{%d}[%s]", len(path), rv.String())
}

// CreateIndexedGridStringBz creates a string of the provided bytes matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringBz[M ~[][]B, B byte | rune, S ~[]E, E XY](vals M, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = string(val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// CreateIndexedGridStringNums creates a string of the provided numbers matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringNums[M ~[][]N, N Integer, S ~[]E, E XY](vals M, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = fmt.Sprintf("%d", val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// CreateIndexedGridStringFunc creates a string of the provided matrix.
// The converter should take in a cell's value and output the string to use for that cell.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridStringFunc[M ~[][]G, G any, S ~[]E, E XY](vals M, converter func(G) string, colorPoints S, highlightPoints S) string {
	strs := make([][]string, len(vals))
	for y, row := range vals {
		strs[y] = make([]string, len(row))
		for x, val := range row {
			strs[y][x] = converter(val)
		}
	}
	return CreateIndexedGridString(strs, colorPoints, highlightPoints)
}

// CreateIndexedGridString creates a string of the provided strings matrix.
// The result will have row and column indexes and the desired cells will be colored and/or highlighted.
func CreateIndexedGridString[S ~[]E, E XY](vals [][]string, colorPoints S, highlightPoints S) string {
	// Get the height. If it's zero, there's nothing to return.
	height := len(vals)
	if height == 0 {
		return ""
	}

	// Get the max cell length and the max row width.
	cellLen := 0
	width := len(vals[0])
	for _, r := range vals {
		if len(r) > width {
			width = len(r)
		}
		for _, c := range r {
			if len(c) > cellLen {
				cellLen = utf8.RuneCountInString(c)
			}
		}
	}
	// Add an extra space if there's two or more characters per cell.
	if cellLen > 1 {
		cellLen++
	}

	// Define the format that each line will start with and for each cell.
	leadFmt := fmt.Sprintf("%%%dd:", len(fmt.Sprintf("%d", height)))
	blankLead := strings.Repeat(" ", len(fmt.Sprintf(leadFmt, 0)))
	cellFmt := fmt.Sprintf("%%%ds", cellLen)

	// If none of the rows have anything, just print out the row numbers.
	if width == 0 {
		lines := make([]string, len(vals))
		for y := range vals {
			lines[y] = fmt.Sprintf(leadFmt, y)
		}
		return strings.Join(lines, "\n")
	}

	// Create the index numbers across the top.
	dCount := len(fmt.Sprintf("%d", width-1))
	dLen := width * cellLen
	digits := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	topIndexLines := make([]string, dCount+1)
	topIndexLines[dCount] = strings.Repeat("-", dLen)
	rep := 1
	for l := 1; l <= dCount; l++ {
		first := " "
		if l == 1 {
			first = "0"
		}
		first = strings.Repeat(fmt.Sprintf(cellFmt, first), rep)

		var sb strings.Builder
		for _, s := range digits {
			if len(first)+sb.Len() >= dLen {
				break
			}
			sb.WriteString(strings.Repeat(fmt.Sprintf(cellFmt, s), rep))
		}

		rep *= 10
		line := first + strings.Repeat(sb.String(), 1+width/rep)
		topIndexLines[dCount-l] = line[:dLen]
	}

	// Create a matrix indicating desired text formats.
	textFmt := make([][]int, height)
	for y := range textFmt {
		textFmt[y] = make([]int, width)
	}
	for _, p := range colorPoints {
		if p.GetY() < height && p.GetX() < width {
			textFmt[p.GetY()][p.GetX()] = 1
		}
	}
	for _, p := range highlightPoints {
		if p.GetY() < height && p.GetX() < width && textFmt[p.GetY()][p.GetX()] <= 1 {
			textFmt[p.GetY()][p.GetX()] += 2
		}
	}

	// Start with the top index lines shifted right a bit to account for row indexes in the lines to follow.
	var rv strings.Builder
	for _, l := range topIndexLines {
		rv.WriteString(fmt.Sprintf("%s%s\n", blankLead, l))
	}

	// Add all the line numbers, and cells (with the desired coloring/marking).
	for y, r := range vals {
		rv.WriteString(fmt.Sprintf(leadFmt, y))
		for x := 0; x < width; x++ {
			v := ""
			if x < len(r) {
				v = r[x]
			}
			cell := fmt.Sprintf(cellFmt, v)
			switch textFmt[y][x] {
			case 0: // default look.
				rv.WriteString(cell)
			case 1: // color only.
				rv.WriteString("\033[94m" + cell + "\033[0m") // Light-blue text.
			case 2: // highlight only
				rv.WriteString("\033[7m" + cell + "\033[0m") // Foreground<->Background Reversed.
			case 3: // color and highlight
				rv.WriteString("\033[94;7m" + cell + "\033[0m") // Light-blue background after fg<->bg reversed.
			default: // Unknown, make it ugly.
				rv.WriteString("\033[93;41m" + cell + "\033[0m") // Bright yellow text on a red background.
			}
		}
		rv.WriteByte('\n')
	}

	return rv.String()
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  CLI params and input parsing  ---------------------------------
// -------------------------------------------------------------------------------------------------

// Params contains anything that might be provided via command-line arguments.
type Params struct {
	// HelpPrinted is whether or not the help message was printed.
	HelpPrinted bool
	// Errors is a list of errors encountered while parsing the arguments.
	Errors []error
	// Count is just a generic int that can be provided.
	Count int
	// Option is another generic int that can be provided.
	Option int
	// InputFile is the file that contains the puzzle data to solve.
	InputFile string
	// Input is the contents of the input file split on newlines.
	Input []string
	// Custom is a set of custom strings to provide as input.
	Custom []string
}

// String creates a multi-line string representing this Params.
func (p Params) String() string {
	nameFmt := "%10s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", debug),
		fmt.Sprintf(nameFmt+"%t", "Verbose", verbose),
		fmt.Sprintf(nameFmt+"%d", "Errors", len(p.Errors)),
		fmt.Sprintf(nameFmt+"%d", "Count", p.Count),
		fmt.Sprintf(nameFmt+"%d", "Option", p.Option),
		fmt.Sprintf(nameFmt+"%s", "Input File", p.InputFile),
		fmt.Sprintf(nameFmt+"%d lines", "Input", len(p.Input)),
		fmt.Sprintf(nameFmt+"%d lines", "Custom", len(p.Custom)),
	}
	if len(p.Errors) > 0 {
		lines = append(lines, fmt.Sprintf("Errors (%d):", len(p.Errors)))
		errors := make([]string, len(p.Errors))
		for i, err := range p.Errors {
			errors[i] = err.Error()
		}
		lines = append(lines, AddLineNumbers(errors, 1)...)
	}
	if len(p.Input) > 0 {
		lines = append(lines, fmt.Sprintf("Input (%d):", len(p.Input)))
		lines = append(lines, AddLineNumbers(p.Input, 0)...)
	}
	if len(p.Custom) > 0 {
		lines = append(lines, fmt.Sprintf("Custom Input (%d):", len(p.Custom)))
		lines = append(lines, AddLineNumbers(p.Custom, 0)...)
	}
	return strings.Join(lines, "\n")
}

// DEFAULT_INPUT_FILE is the default input filename.
const DEFAULT_INPUT_FILE = "example.input"

// GetParams parses the provided args into the command's params.
func GetParams(args []string) *Params {
	defer FuncEnding(FuncStarting())
	var err error
	rv := Params{}
	countGiven := false
	verboseGiven := false
	for i := 0; i < len(args); i++ {
		switch {
		// Flag cases go first.
		case IsOneOfStrFold(args[i], "--help", "-h", "help"):
			Debugf("Help flag found: [%s].", args[i])
			lines := []string{
				fmt.Sprintf("Usage: %s [<input file>] [<flags>]", GetMyExe()),
				fmt.Sprintf("Default <input file> is %s", DEFAULT_INPUT_FILE),
				"Flags:",
				"  --debug       Turns on debugging.",
				"  --verbose|-v  Turns on verbose output.",
				"",
				"Single Options:",
				"  Providing these multiple times will overwrite the previously provided value.",
				"  --input|-i <input file>  An option to define the input file.",
				"  --count|-n <number>      Defines a count.",
				"  --option|-o <number>     Defines an option value.",
				"",
				"Repeatable Options:",
				"  Providing these multiple times will add to previously provided values.",
				"  Values are read until the next one starts with a dash.",
				"  To provide entries that start with a dash, you can use --flag='<value>' syntax.",
				"  --lines|-l <value 1> [<value 2> ...]  Defines custom input lines.",
				"",
			}
			// Not using Stdoutf() here because the extra formatting is annoying with help text.
			fmt.Println(strings.Join(lines, "\n"))
			rv.HelpPrinted = true
		case HasOneOfPrefixesFold(args[i], "--debug", "-vv"):
			Debugf("Debug option found: [%s], args left: %q.", args[i], args[i:])
			var extraI int
			oldDebug := debug
			debug, extraI, err = ParseFlagBool(args[i:])
			i += extraI
			rv.AppendError(err)
			if err == nil {
				switch {
				case !oldDebug && debug:
					Stderrf("Debugging enabled by CLI arguments.")
				case oldDebug && !debug:
					Stderrf("Debugging disabled by CLI arguments.")
				}
			}
		case HasOneOfPrefixesFold(args[i], "--verbose", "-v"):
			Debugf("Verbose option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			verbose, extraI, err = ParseFlagBool(args[i:])
			i += extraI
			rv.AppendError(err)
			verboseGiven = true
		case HasOneOfPrefixesFold(args[i], "--input", "--input-file"):
			Debugf("Input file option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			rv.InputFile, extraI, err = ParseFlagString(args[i:])
			i += extraI
			rv.AppendError(err)
		case HasOneOfPrefixesFold(args[i], "--count", "-c", "-n"):
			Debugf("Count option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			rv.Count, extraI, err = ParseFlagInt(args[i:])
			i += extraI
			rv.AppendError(err)
			countGiven = true
		case HasOneOfPrefixesFold(args[i], "--option", "--opt", "-o"):
			Debugf("Option option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			rv.Option, extraI, err = ParseFlagInt(args[i:])
			i += extraI
			rv.AppendError(err)
		case HasOneOfPrefixesFold(args[i], "--line", "--lines", "-l", "--custom", "--val"):
			Debugf("Custom option found: [%s], args after: %q.", args[i], args[i:])
			var extraI int
			var vals []string
			vals, extraI, err = ParseRepeatedFlagString(args[i:])
			rv.Custom = append(rv.Custom, vals...)
			i += extraI
			rv.AppendError(err)

		// Positional args go last in the order they're expected.
		case len(rv.InputFile) == 0 && len(args[i]) > 0 && args[i][0] != '-':
			Debugf("Input File argument: [%s], args after: %q", args[i], args[i:])
			rv.InputFile = args[i]
		default:
			Debugf("Unknown argument found: [%s], args after: %q.", args[i], args[i:])
			rv.AppendError(fmt.Errorf("unknown argument %d: [%s]", i+1, args[i]))
		}
	}
	if len(rv.InputFile) == 0 {
		rv.InputFile = DEFAULT_INPUT_FILE
	}
	if !verboseGiven {
		verbose = debug
	}
	if !countGiven {
		rv.Count = DEFAULT_COUNT
	}
	return &rv
}

// AppendError adds an error to this Params as long as the error is not nil.
func (p *Params) AppendError(err error) {
	if err != nil {
		p.Errors = append(p.Errors, err)
	}
}

// HasError returns true if this Params has one or more errors.
func (p Params) HasError() bool {
	return len(p.Errors) != 0
}

// Error flattens the Errors slice into a single string.
// It also makes the Params struct satisfy the error interface.
func (p Params) GetError() error {
	switch len(p.Errors) {
	case 0:
		return nil
	case 1:
		return p.Errors[0]
	default:
		errs := make([]error, 1, 1+len(p.Errors))
		errs[0] = fmt.Errorf("Found %d errors:", len(p.Errors)) //nolint:stylecheck,revive // punct okay here.
		for i, err := range p.Errors {
			errs = append(errs, fmt.Errorf("  %d: %w", i+1, err))
		}
		return errors.Join(errs...)
	}
}

// HasCustom returns true if the provided string was given as a custom arg.
func (p Params) HasCustom(str string) bool {
	for _, cust := range p.Custom {
		if cust == str {
			return true
		}
	}
	return false
}

// IsOneOfStrFold tests if the given string is equal (ignoring case) to one of the given options.
func IsOneOfStrFold(str string, opts ...string) bool {
	for _, opt := range opts {
		if strings.EqualFold(str, opt) {
			return true
		}
	}
	return false
}

// HasPrefixFold tests if the given string starts with the given prefix (ignoring case).
func HasPrefixFold(str, prefix string) bool {
	return len(str) >= len(prefix) && strings.EqualFold(str[0:len(prefix)], prefix)
}

// HasOneOfPrefixesFold tests if the given string has one of the given prefixes.
func HasOneOfPrefixesFold(str string, prefixes ...string) bool {
	for _, pre := range prefixes {
		if HasPrefixFold(str, pre) {
			return true
		}
	}
	return false
}

// ParseBool converts a string into a bool.
// First return bool is the parsed value.
// Second return bool is whether or not the parsing was successful.
func ParseBool(str string) (val bool, isBool bool) {
	// Note: Not using strconv.ParseBool because I want it a bit looser (any casing) and to allow yes/no/off/on values.
	lstr := strings.ToLower(strings.TrimSpace(str))
	switch lstr {
	case "false", "f", "0", "no", "n", "off":
		isBool = true
	case "true", "t", "1", "yes", "y", "on":
		val = true
		isBool = true
	}
	return
}

// ParseFlagString parses a string flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's string value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseFlagString(args []string) (string, int, error) {
	if strings.ContainsAny(args[0], "= ") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(args[0], " ", 2)
		}
		if len(parts) == 2 {
			if len(parts[1]) > 1 {
				for _, c := range []string{`'`, `"`} {
					if parts[1][:1] == c && parts[1][len(parts[1])-1:] == c {
						return parts[1][1 : len(parts[1])-1], 0, nil
					}
				}
			}
			return parts[1], 0, nil
		}
		return "", 0, fmt.Errorf("unable to split flag and value from string: [%s]", args[0])
	}
	if len(args) > 1 {
		return args[1], 1, nil
	}
	return "", 0, fmt.Errorf("no value provided after %s flag", args[0])
}

// ParseRepeatedFlagString parses a flag that allows providing multiple strings.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's string value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseRepeatedFlagString(args []string) ([]string, int, error) {
	if strings.ContainsAny(args[0], "= ") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(args[0], " ", 2)
		}
		if len(parts) != 2 {
			return []string{}, 0, fmt.Errorf("unable to split flag and value from string: [%s]", args[0])
		}
		if len(parts[1]) > 1 {
			for _, c := range []string{`'`, `"`} {
				if parts[1][:1] == c && parts[1][len(parts[1])-1:] == c {
					parts[1] = parts[1][1 : len(parts[1])-1]
				}
			}
		}
		return parts[1:], 0, nil
	}
	rv := []string{}
	for _, arg := range args[1:] {
		if arg[0] == '-' {
			return rv, len(rv), nil
		}
		rv = append(rv, arg)
	}
	if len(rv) > 0 {
		return rv, len(rv), nil
	}
	return rv, 0, fmt.Errorf("no values provided after %s flag", args[0])
}

// ParseFlagBool parses a boolean flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and parsed.
// Otherwise, if args[1] is a boolean string value, that is parsed.
// Otherwise, the flag defaults to true.
//
// The first return value is the parsed boolean value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseFlagBool(args []string) (bool, int, error) {
	if strings.ContainsAny(args[0], "= ") {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) == 1 {
			parts = strings.SplitN(args[0], " ", 2)
		}
		if len(parts) == 2 {
			val, isBool := ParseBool(parts[1])
			if !isBool {
				return false, 0, fmt.Errorf("invalid %s bool value: [%s]", parts[0], parts[1])
			}
			return val, 0, nil
		}
		return false, 0, fmt.Errorf("unable to split flag and value from string: [%s]", args[0])
	}
	if len(args) > 1 {
		val, isBool := ParseBool(args[1])
		if isBool {
			return val, 1, nil
		}
	}
	return true, 0, nil
}

// ParseFlagInt parses an int flag from arguments.
//
// The flag in question should be in args[0].
// If args[0] contains "=" or " " then the desired value will be extracted from that string and returned.
// Otherwise, if args[1] exists, that is returned.
// Otherwise, an error is given.
//
// The first return value is the flag's int value.
// The second return value is the number of extra arguments used.
// The third return value is any error encountered.
func ParseFlagInt(args []string) (int, int, error) {
	rvStr, used, err := ParseFlagString(args)
	if err != nil {
		return 0, used, err
	}
	var rv int
	rv, err = strconv.Atoi(rvStr)
	if err != nil {
		return 0, used, err
	}
	return rv, used, nil
}

// ReadFile reads a file and splits it into lines.
func ReadFile(filename string) ([]string, error) {
	defer FuncEndingAlways(FuncStarting(filename))
	DebugAlwaysf("Reading file: %s", filename)
	dat, err := os.ReadFile(filename)
	if err != nil {
		Stderrf("error reading file: %v", err)
		return []string{}, err
	}
	rv := strings.Split(string(dat), "\n")
	for len(rv[len(rv)-1]) == 0 {
		rv = rv[:len(rv)-1]
	}
	return rv, nil
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  Environment Variable Handling  --------------------------------
// -------------------------------------------------------------------------------------------------

// HandleEnvVars looks at specific environment variables and sets global variables appropriately.
func HandleEnvVars() error {
	var err error
	debug, err = GetEnvVarBool("DEBUG")
	if debug {
		Stderrf("Debugging enabled via environment variable.")
	}
	return err
}

// GetEnvVarBool gets the environment variable with the given name and converts it to a bool.
func GetEnvVarBool(name string) (bool, error) {
	str := os.Getenv(name)
	if len(str) == 0 {
		return false, nil
	}
	val, isBool := ParseBool(str)
	if !isBool {
		return false, fmt.Errorf("invalid %s env var boolean value: [%s]", name, str)
	}
	return val, nil
}

// -------------------------------------------------------------------------------------------------
// ------------------------------  Function start/stop timing stuff  -------------------------------
// -------------------------------------------------------------------------------------------------

// If all you want is starting/ending messages when debug is on, use:
//    defer FuncEnding(FuncStarting())
// If, when debug is on, you want starting/ending messages,
// but when debug is off, you still want the function duration, then use:
//    defer FuncEndingAlways(FuncStarting())

// FuncStarting outputs that a function is starting (if debug is true).
// It returns the params needed by FuncEnding or FuncEndingAlways.
//
// Arguments provided will be converted to stings using %v and included as part of the function name.
// Only provide minimal values needed to differentiate start/stop output lines.
// Long strings and complex structs should be avoided as args.
//
// Example 1: In a function named "foo", you have this:
//
//	  FuncStarting()
//	The printed message will note that "foo" is starting.
//	That same string will also be returned as the 2nd return paremeter.
//
// Example 2: In a function named "bar", you have this:
//
//	  FuncStarting(3 * time.Second)
//	The printed message will note that "bar: 3s" is starting.
//	That same string will also be returned as the 2nd return paremeter.
//
// Example 3:
//
//	  func sum(ints ...int) {
//	      FuncStarting(ints...)
//	  }
//	  sum(1, 2, 3, 4, 20, 21, 22)
//	The printed message will note that "sum: 1, 2, 3, 4, 20, 21, 22" is starting.
//	That same string will also be returned as the 2nd return paremeter.
//
// Standard Usage: defer FuncEnding(FuncStarting())
//
//	Or: defer FuncEndingAlways(FuncStarting())
func FuncStarting(a ...interface{}) (time.Time, string) {
	funcDepth++
	name := GetFuncName(1, a...)
	DebugAsf(name, "Starting.")
	return time.Now(), name
}

// FuncStartingAlways is the same as FuncStarting except if debug is off, output will go to stdout.
//
// This differs from FuncStarting in that this will always do the output (regardless of debug state).
//
// Usage: defer FuncEndingAlways(FuncStartingAlways())
func FuncStartingAlways(a ...interface{}) (time.Time, string) {
	funcDepth++
	name := GetFuncName(1, a...)
	DebugAlwaysAsf(name, "Starting.")
	return time.Now(), name
}

const DONE_FMT = "Done. Duration: [%s]."

var panicPrinted bool

// FuncEnding decrements the function depth and, if debug is on, outputs to stderr how long a function took.
// Args will usually come from FuncStarting().
//
// This differs from FuncEndingAlways in that this only outputs something if debugging is turned on.
//
// Usage: defer FuncEnding(FuncStarting())
func FuncEnding(start time.Time, name string) {
	if !panicPrinted {
		if r := recover(); r != nil {
			DebugAlwaysAsf(name, "PANIC")
			panicPrinted = true
			defer func() {
				panic(r)
			}()
		}
	}
	if !panicPrinted {
		DebugAsf(name, DONE_FMT, time.Since(start))
	}
	if funcDepth > -1 {
		funcDepth--
	}
}

// FuncEndingAlways is the same as FuncEnding except if debug is off, output will go to stdout.
//
// This differs from FuncEnding in that this will always do the output (regardless of debug state).
//
// Usage: defer FuncEndingAlways(FuncStarting())
func FuncEndingAlways(start time.Time, name string) {
	if !panicPrinted {
		if r := recover(); r != nil {
			DebugAlwaysAsf(name, "PANIC")
			panicPrinted = true
			defer func() {
				panic(r)
			}()
		}
	}
	if !panicPrinted {
		DebugAlwaysAsf(name, DONE_FMT, time.Since(start))
	}
	if funcDepth > -1 {
		funcDepth--
	}
}

// DurClock converts a duration to a string in minimal clock notation with nanosecond precision.
//
// - If one or more hours, format is "H:MM:SS.NNNNNNNNNs", e.g. "12:01:02.000000000".
// - If less than one hour, format is "M:SS.NNNNNNNNNs",   e.g. "34:00.000000789".
// - If less than one minute, format is "S.NNNNNNNNNs",    e.g. "56.000456000".
// - If less than one second, format is "0.NNNNNNNNNs",    e.g. "0.123000000".
func DurClock(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes())
	s := int(d.Seconds())
	n := int(d.Nanoseconds()) - 1000000000*s
	s -= 60 * m
	m -= 60 * h
	switch {
	case h > 0:
		return fmt.Sprintf("%d:%02d:%02d.%09d", h, m, s, n)
	case m > 0:
		return fmt.Sprintf("%d:%02d.%09d", m, s, n)
	default:
		return fmt.Sprintf("%d.%09d", s, n)
	}
}

// GetFuncName gets the name of the function at the given depth.
//
// Depth 0 = the function calling GetFuncName.
// Depth 1 = the function calling the function calling GetFuncName.
// Etc.
//
// Extra arguments provided will be converted to stings using %v and included as part of the function name.
// Only values needed to differentiate start/stop output lines should be provided.
// Long strings and complex structs should be avoided.
func GetFuncName(depth int, a ...interface{}) string {
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, more := frames.Next()
	for more && depth > 0 {
		frame, more = frames.Next()
		depth--
	}
	name := strings.TrimPrefix(frame.Function, "main.")
	// Using a switch to prevent calling strings.Join for small (common) use cases. Saves a little mem and processing.
	switch len(a) {
	case 0:
		// do nothing
	case 1:
		name += fmt.Sprintf(": %v", a[0])
	case 2:
		name += fmt.Sprintf(": %v, %v", a[0], a[1])
	case 3:
		name += fmt.Sprintf(": %v, %v, %v", a[0], a[1], a[2])
	default:
		args := make([]string, len(a))
		for i, arg := range a {
			args[i] = fmt.Sprintf("%v", arg)
		}
		name += fmt.Sprintf(": %s", strings.Join(args, ", "))
	}
	return name
}

// GetMyExe returns how to execute this program by parsing os.Args[0].
func GetMyExe() string {
	_, name := filepath.Split(os.Args[0])
	if i := strings.Index(os.Args[0], "/go-build"); i == -1 {
		name = "./" + name
	} else {
		name = fmt.Sprintf("go run %s.go", name)
	}
	return name
}

// -------------------------------------------------------------------------------------------------
// ---------------------------------------  Output wrappers  ---------------------------------------
// -------------------------------------------------------------------------------------------------

// GetOutputPrefix gets the prefix to add to all output.
func GetOutputPrefix(funcName string) string {
	tabs := ""
	if debug && funcDepth > 0 {
		tabs = strings.Repeat("  ", funcDepth)
	}
	return fmt.Sprintf("(%14s) %s[%s] ", DurClock(time.Since(startTime)), tabs, funcName)
}

// Stdoutf outputs to stdout with a prefixed run duration and automatic function name.
func Stdoutf(format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// Stderrf outputs to stderr with a prefixed run duration and automatic function name.
func Stderrf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// StdoutAsf outputs to stdout with a prefixed run duration and provided function name.
func StdoutAsf(funcName, format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(funcName)+format+"\n", a...)
}

// StderrAsf outputs to stderr with a prefixed run duration and provided function name.
func StderrAsf(funcName, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(funcName)+format+"\n", a...)
}

// Debugf is like Stderrf if the debug flag is set; otherwise it does nothing.
func Debugf(format string, a ...interface{}) {
	if debug {
		StderrAsf(GetFuncName(1), format, a...)
	}
}

// DebugAsf is like StderrAsf if the debug flag is set; otherwise it does nothing.
func DebugAsf(funcName, format string, a ...interface{}) {
	if debug {
		StderrAsf(funcName, format, a...)
	}
}

// DebugAlwaysf is like Stderrf if the debug flag is set; otherwise it's like Stdoutf.
func DebugAlwaysf(format string, a ...interface{}) {
	if debug || verbose {
		StderrAsf(GetFuncName(1), format, a...)
	} else {
		StdoutAsf(GetFuncName(1), format, a...)
	}
}

// DebugAlwaysAsf is like StderrAsf if the debug flag is set; otherwise it's like StdoutAsf.
func DebugAlwaysAsf(funcName, format string, a ...interface{}) {
	if debug || verbose {
		StderrAsf(funcName, format, a...)
	} else {
		StdoutAsf(funcName, format, a...)
	}
}

// Verbosef outputs to Stderr if the verbose flag was provided. Does nothing otherwise.
func Verbosef(format string, a ...interface{}) {
	if verbose {
		StderrAsf(GetFuncName(1), format, a...)
	}
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  Primary Program Running Parts  --------------------------------
// -------------------------------------------------------------------------------------------------

var (
	// Debug is a flag for whether or not debug messages should be displayed.
	debug bool
	// Verbose is a flag for whether or not verbose messages should be displayed.
	verbose bool
)

// startTime is the time when the program started.
var startTime time.Time

// funcDepth is a global counter keeping track of function depth by the starting/ending function functions.
var funcDepth int

func init() {
	funcDepth = -1
}

// main is the main function that gets run for this file.
func main() {
	startTime = time.Now()
	// Handle the env vars before calling into Run().
	// That way, if debug is on, we will get the start message for Run().
	err := HandleEnvVars()
	if err == nil {
		err = Run()
	}
	if err != nil {
		// Not using Stderrf(...) here because I don't want the time and function prefix on this.
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run does all the primary coordination for this program.
// It's basically a replacement for main() that returns an error.
func Run() error {
	defer FuncEndingAlways(FuncStarting())
	params := GetParams(os.Args[1:])
	if params.HelpPrinted {
		return nil
	}
	if !params.HasError() {
		var err error
		params.Input, err = ReadFile(params.InputFile)
		params.AppendError(err)
	}
	Debugf("Params:\n%s", params)
	if params.HasError() {
		return params.GetError()
	}
	answer, err := Solve(params)
	if err != nil {
		return err
	}
	Stdoutf("Answer: %s", answer)
	return nil
}
