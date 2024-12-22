package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const DEFAULT_COUNT = 25

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	Debugf("Parsed Input:\n%s", input)
	InitStuff()
	if debug {
		Stderrf("%s", NumPad)
		Stderrf("%s", DirPad)
	}

	topBot := NewRobot("Bottom", NumPad, nil)
	for i := 0; i < params.Count; i++ {
		topBot = NewRobot(fmt.Sprintf("Bot %d", i+1), DirPad, topBot)
	}

	answer := 0
	for _, code := range input.Codes {
		result := topBot.EnterCode(code.Keys)
		v1 := len(result)
		v2 := code.Value
		a := v1 * v2
		Verbosef("%s: %d = %d * %d: %s", string(code.Keys), a, v1, v2, result)
		answer += a
	}

	return fmt.Sprintf("%d", answer), nil
}

func EnterSequence(robot *Robot, seq Sequence) string {
	if debug {
		Stderrf("[%d/%d]: %s\n%s", 0, len(seq), seq, DrawRobots(robot))
	}
	var rv string
	for i, step := range seq {
		if debug {
			var l, r string
			if i > 0 {
				l = string(seq[:i])
			}
			if i < len(seq)-1 {
				r = string(seq[i+1:])
			}
			Debugf("[%d/%d]: %s  %c  %s", i+1, len(seq), l, step, r)
		}
		res := robot.TapKey(step)
		if debug {
			Stderrf("[%d/%d]:\n%s", i+1, len(seq), DrawRobots(robot))
		}
		if res != NoResult {
			rv += string(res)
			Debugf("Got a key back: %c", res)
		}
	}
	return rv
}

type Robot struct {
	Name   string
	Cur    byte
	Keypad *Keypad
	Next   *Robot
}

func NewRobot(name string, keypad *Keypad, next *Robot) *Robot {
	return &Robot{Name: name, Cur: A, Keypad: keypad, Next: next}
}

func (r *Robot) EnterCode(buttons []byte) Sequence {
	var rv Sequence
	for _, key := range buttons {
		opts := r.PushButton(key)
		Debugf("%s: To press '%c': %s", r.Name, key, StringNumberJoin(opts, 1, "  "))
		var best Sequence
		for _, seq := range opts {
			if len(best) == 0 || len(seq) < len(best) {
				best = seq
			}
		}
		rv = append(rv, best...)
	}
	Debugf("On %s, to type %q: %s", r.Name, string(buttons), rv)
	return rv
}

func (r *Robot) PushButton(toPush byte) []Sequence {
	if r.Next == nil {
		rv := r.Keypad.Paths[r.Cur][toPush]
		r.Cur = toPush
		return rv
	}

	nexts := r.Next.PushButton(toPush)
	rv := make([]Sequence, len(nexts))
	for i, next := range nexts {
		rv[i] = ExpandDirSequence(r.Cur, next)
	}

	return rv
}

func (r *Robot) TapKey(key byte) byte {
	if key == A {
		if r.Next == nil {
			return r.Cur
		}
		return r.Next.TapKey(r.Cur)
	}
	curP := r.Keypad.Points[r.Cur]
	d := DirDelta[key]
	c := AddPoints(curP, d)
	r.Cur = r.Keypad.Layout[c.Y][c.X]
	if r.Cur == Empty {
		panic("moved onto empty space")
	}
	return NoResult
}

func (r *Robot) Reset() {
	r.Cur = A
	if r.Next != nil {
		r.Next.Reset()
	}
}

func (r *Robot) DrawKeypad() []string {
	lines := r.Keypad.Draw(r.Cur)
	rv := []string{PadCenter(r.Name, len(lines[0]))}
	rv = append(rv, lines...)
	return rv
}

func DrawRobots(robot *Robot) string {
	var keypads [][]string
	maxHeight := 0
	cur := robot
	for cur != nil {
		kp := cur.DrawKeypad()
		if len(kp) > maxHeight {
			maxHeight = len(kp)
		}
		keypads = append(keypads, kp)
		cur = cur.Next
	}

	lines := make([]string, maxHeight)
	sep := ""
	for _, kp := range keypads {
		blank := strings.Repeat(" ", len(kp[0]))
		for i := 0; i < maxHeight; i++ {
			if i < len(kp) {
				lines[i] += sep + kp[i]
			} else {
				lines[i] += sep + blank
			}
		}
		sep = "   "
	}

	return strings.Join(lines, "\n")
}

func PadCenter(str string, width int) string {
	if len(str) >= width {
		return str
	}
	toAdd := width - len(str)
	addLeft := toAdd / 2
	addRight := toAdd - addLeft
	if addLeft < addRight {
		addLeft, addRight = addRight, addLeft
	}
	return strings.Repeat(" ", addLeft) + str + strings.Repeat(" ", addRight)
}

const (
	A        = byte('A')
	Up       = byte('^')
	Down     = byte('v')
	Left     = byte('<')
	Right    = byte('>')
	Empty    = byte(' ')
	NoResult = byte('x')
)

var (
	Dirs = []byte{Up, Down, Left, Right}

	DUp    = NewPoint(0, -1)
	DDown  = NewPoint(0, 1)
	DLeft  = NewPoint(-1, 0)
	DRight = NewPoint(1, 0)

	DirDelta = map[byte]*Point{
		Up:    DUp,
		Down:  DDown,
		Left:  DLeft,
		Right: DRight,
	}
)

var (
	NumPadLayout = [][]byte{
		{'7', '8', '9'},
		{'4', '5', '6'},
		{'1', '2', '3'},
		{Empty, '0', A},
	}

	// Populated with InitStuff().
	NumPadButtons []byte
	NumPadPoints  map[byte]*Point
	NumPadPaths   map[byte]map[byte][]Sequence
	NumPadNexts   map[byte]map[byte]byte

	DirPadLayout = [][]byte{
		{Empty, Up, A},
		{Left, Down, Right},
	}

	DirPadNexts = map[byte]map[byte]byte{
		Up:    {Right: A, Down: Down},
		A:     {Left: Up, Down: Right},
		Left:  {Right: Down},
		Down:  {Left: Left, Up: Up, Right: Right},
		Right: {Up: A, Left: Down},
	}

	// Hard coding this one because it's small enough and I'm tired.
	DirPadPaths = map[byte]map[byte][]Sequence{
		Up: {
			Up:    {{A}},
			A:     {{Right, A}},
			Left:  {{Down, Left, A}},
			Down:  {{Down, A}},
			Right: {{Right, Down, A}, {Down, Right, A}},
		},
		A: {
			Up:    {{Left, A}},
			A:     {{A}},
			Left:  {{Down, Left, Left, A}}, // {Left, Down, Left, A} is never better.
			Down:  {{Left, Down, A}, {Down, Left, A}},
			Right: {{Down, A}},
		},
		Left: {
			Up:    {{Right, Up, A}},
			A:     {{Right, Right, Up, A}}, // {Right, Up, Right, A is never better.
			Left:  {{A}},
			Down:  {{Right, A}},
			Right: {{Right, Right, A}},
		},
		Down: {
			Up:    {{Up, A}},
			A:     {{Up, Right, A}, {Right, Up, A}},
			Left:  {{Left, A}},
			Down:  {{A}},
			Right: {{Right, A}},
		},
		Right: {
			Up:    {{Up, Left, A}, {Left, Up, A}},
			A:     {{Up, A}},
			Left:  {{Left, Left, A}},
			Down:  {{Left, A}},
			Right: {{A}},
		},
	}

	// Populated with InitStuff().
	DirPadButtons []byte
	DirPadPoints  map[byte]*Point
)

var (
	NumPad *Keypad
	DirPad *Keypad
)

type Keypad struct {
	Name    string
	Layout  [][]byte
	Buttons []byte
	Points  map[byte]*Point
	Nexts   map[byte]map[byte]byte
	Paths   map[byte]map[byte][]Sequence
}

func NewKeypad(name string, layout [][]byte) *Keypad {
	rv := &Keypad{
		Name:    name,
		Layout:  layout,
		Buttons: make([]byte, 0, len(layout)*len(layout[0])-1),
		Points:  make(map[byte]*Point),
		Nexts:   make(map[byte]map[byte]byte),
		Paths:   make(map[byte]map[byte][]Sequence),
	}

	for y := range layout {
		for x, v := range layout[y] {
			if v == Empty {
				continue
			}
			rv.Buttons = append(rv.Buttons, v)
			rv.Points[v] = NewPoint(x, y)
			p := NewPoint(x, y)
			rv.Nexts[v] = make(map[byte]byte)
			for dir, d := range DirDelta {
				n, ok := GetB(layout, AddPoints(p, d))
				if !ok || n == Empty {
					continue
				}
				rv.Nexts[v][dir] = n
			}
		}
	}
	slices.Sort(rv.Buttons)

	for _, from := range rv.Buttons {
		rv.Paths[from] = make(map[byte][]Sequence)
		fromP := rv.Points[from]
		for _, to := range rv.Buttons {
			toP := rv.Points[to]
			seqs := DeltaToSequences(toP.Sub(fromP))
			seqs = FilterIllegal(seqs, fromP, layout)
			seqs = ReduceSequences(seqs)
			rv.Paths[from][to] = seqs
		}
	}
	ValidatePaths(name, rv.Paths, rv.Points)

	return rv
}

func (k *Keypad) String() string {
	layout := CreateIndexedGridStringFunc(k.Layout, GetButtonName, []*Point(nil), nil)
	buttons := strings.Join(MapSlice(k.Buttons, GetButtonName), " ")
	points := PointMapString(k.Points)
	var paths []string
	nexts := make([]string, 0, len(k.Buttons))
	for _, from := range k.Buttons {
		for _, to := range k.Buttons {
			seqs := k.Paths[from][to]
			paths = append(paths, fmt.Sprintf("%c to %c: (%d) %s", from, to, len(seqs), StringNumberJoin(seqs, 1, " ")))
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%c: {", from))
		first := true
		for _, dir := range Dirs {
			next, ok := k.Nexts[from][dir]
			if !ok {
				continue
			}
			if !first {
				sb.WriteString(", ")
			}
			first = false
			sb.WriteString(fmt.Sprintf("%c=%c", dir, next))
		}
		sb.WriteByte('}')
		nexts = append(nexts, sb.String())
	}
	return fmt.Sprintf("%s\nLayout:\n%sButtons: %s\nPoints: %s\nPaths:\n%s\nNexts:\n%s",
		k.Name, layout, buttons, points, strings.Join(paths, "\n"), strings.Join(nexts, "\n"))
}

const (
	hDivLong  = "+---+---+---+"
	hDivShort = "    +---+---+"
)

func (k *Keypad) Draw(cur byte) []string {
	lines := make([]string, 0, len(k.Layout)*2+1)
	for y := range k.Layout {
		if y != 0 || k.Layout[0][0] != Empty {
			lines = append(lines, hDivLong)
		} else {
			lines = append(lines, hDivShort)
		}
		line := ""
		for _, v := range k.Layout[y] {
			switch {
			case v == Empty:
				line += "    "
			case v == cur:
				line += fmt.Sprintf("|\033[7m %c \033[0m", v)
			default:
				line += fmt.Sprintf("| %c ", v)
			}
		}
		line += "|"
		lines = append(lines, line)
	}
	if k.Layout[len(k.Layout)-1][0] == Empty {
		lines = append(lines, hDivShort)
	} else {
		lines = append(lines, hDivLong)
	}
	return lines
}

func InitStuff() {
	DirPad = NewKeypad("Direction Pad", DirPadLayout)
	DirPadButtons = DirPad.Buttons
	DirPadPoints = DirPad.Points
	DirPadNexts = DirPad.Nexts
	DirPadPaths = DirPad.Paths

	NumPad = NewKeypad("Number Pad", NumPadLayout)
	NumPadButtons = NumPad.Buttons
	NumPadPoints = NumPad.Points
	NumPadNexts = NumPad.Nexts
	NumPadPaths = NumPad.Paths
}

func ValidatePaths(name string, paths map[byte]map[byte][]Sequence, keypadPoints map[byte]*Point) {
	for from := range paths {
		for to, seqs := range paths[from] {
			for i, seq := range seqs {
				if len(seq) == 0 {
					Stderrf("%s[%c][%c][%d]: %s", name, from, to, i, seq)
					panic("dir-pad sequence is empty")
				}
				if seq[len(seq)-1] != A {
					Stderrf("%s[%c][%c][%d]: %s", name, from, to, i, seq)
					panic("sequence does not end with A")
				}

				fromP := keypadPoints[from]
				toP := keypadPoints[to]
				check := fromP.Follow(seq)
				if !IsSameXY(check, toP) {
					Stderrf("%s[%c][%c][%d]: %s", name, from, to, i, seq)
					Stderrf("   Start: %s", fromP)
					Stderrf("Expected: %s", toP)
					Stderrf("  Actual: %s", check)
					panic("sequence does not end at expected point")
				}
			}
		}
	}
	Debugf("%s paths are valid.", name)
}

func GetButtonName(button byte) string {
	if button == Empty {
		return "   "
	}
	return fmt.Sprintf(" %c ", button)
}

func PointMapString(pointMap map[byte]*Point) string {
	parts := make([]string, 0, len(pointMap))
	for _, key := range slices.Sorted(maps.Keys(pointMap)) {
		parts = append(parts, fmt.Sprintf("[%d]=%s", key, pointMap[key]))
	}
	return strings.Join(parts, " ")
}

func DeltaToSequences(d *Point) []Sequence {
	if d.X == 0 && d.Y == 0 {
		return []Sequence{{A}}
	}

	yCount := d.Y
	yDir := Down
	if d.Y < 0 {
		yCount *= -1
		yDir = Up
	}
	xCount := d.X
	xDir := Right
	if d.X < 0 {
		xCount *= -1
		xDir = Left
	}

	if xCount == 0 {
		newSeq := RepSeq(yDir, yCount+1)
		newSeq[yCount] = A
		return []Sequence{newSeq}
	}

	if yCount == 0 {
		newSeq := RepSeq(xDir, xCount+1)
		newSeq[xCount] = A
		return []Sequence{newSeq}
	}

	fullSeqX := RepSeq(xDir, xCount)
	fullSeqY := RepSeq(yDir, yCount)
	var rv []Sequence

	// Make all the sequences that move vertically first
	for y := yCount; y > 0; y-- {
		newSeq := RepSeq(yDir, xCount+yCount+1)
		copy(newSeq[y:], fullSeqX)
		newSeq[len(newSeq)-1] = A
		rv = append(rv, newSeq)
	}

	// And now the ones that move horizontally first.
	for x := xCount; x > 0; x-- {
		newSeq := RepSeq(xDir, xCount+yCount+1)
		copy(newSeq[x:], fullSeqY)
		newSeq[len(newSeq)-1] = A
		rv = append(rv, newSeq)
	}

	return rv
}

func FilterIllegal(seqs []Sequence, from *Point, numpad [][]byte) []Sequence {
	if v, ok := GetB(numpad, from); !ok || v == Empty {
		return nil
	}

	rv := make([]Sequence, 0, len(seqs))
seqLoop:
	for _, seq := range seqs {
		if len(seq) == 0 {
			// Debugf("Omitting: Sequence is empty.")
			continue
		}
		if len(seq) == 1 && seq[0] == A {
			// Debugf("Keeping: No-move sequence: %s", seq)
			rv = append(rv, seq)
			continue
		}

		cur := from
		for i, step := range seq {
			if step == A {
				continue
			}

			cur = MovePoint(cur, step)
			v, isIn := GetB(numpad, cur)
			if !isIn {
				_ = i // Debugf("Omitting: From %s, sequence %s goes out of bounds at step %d = %s", from, seq, i+1, cur)
				continue seqLoop
			}
			if v == Empty {
				// Debugf("Omitting: From %s, sequence %s goes to empty space at step %d = %s", from, seq, i+1, cur)
				continue seqLoop
			}
		}

		// Debugf("Keeping: Valid sequence: %s", seq)
		rv = append(rv, seq)
	}

	return rv
}

func ReduceSequences(seqs []Sequence) []Sequence {
	// This is only designed for the initial sequences.
	if len(seqs) <= 1 {
		return seqs
	}

	byDir := make(map[byte][]Sequence)
	for _, seq := range seqs {
		byDir[seq[0]] = append(byDir[seq[0]], seq)
	}

	rv := make([]Sequence, 0, 1+len(byDir)*2)
	if len(byDir[A]) > 0 {
		rv = append(rv, byDir[A][0])
	}

	for _, dir := range Dirs {
		if len(byDir[dir]) == 0 {
			continue
		}
		if len(byDir[dir]) == 1 {
			rv = append(rv, byDir[dir][0])
			continue
		}

		hist := make(map[int][]Sequence)
		best := MAX_INT
		for s, seq := range byDir[dir] {
			seq2 := ExpandDirSequence(seq[0], seq)
			dist := len(seq2)
			_ = s // Debugf("%c[%d] seq %s (%d) expands to %s (%d)", dir, s, seq, len(seq), seq2, dist)
			hist[dist] = append(hist[dist], seq)
			if dist < best {
				best = dist
			}
		}

		rv = append(rv, hist[best][0])
	}
	return rv
}

func HistMapString(hist map[int][]Sequence) string {
	parts := make([]string, 0, len(hist)*2)
	for _, key := range slices.Sorted(maps.Keys(hist)) {
		parts = append(parts, fmt.Sprintf("%4d: %s", key, StringNumberJoin(hist[key], 1, "  ")))
	}
	return strings.Join(parts, "\n")
}

var known = make(map[byte]map[string]Sequence)

func ExpandDirSequence(cur byte, seq Sequence) Sequence {
	if len(seq) == 0 {
		return nil
	}
	seqKey := string(seq)
	curKey := cur
	if known[curKey] != nil && known[curKey][seqKey] != nil {
		return known[curKey][seqKey]
	}

	rv := make(Sequence, 0, len(seq)*2)
	for _, step := range seq {
		next := DirPadPaths[cur][step][0]
		rv = append(rv, next...)
		cur = step
	}

	if known[curKey] == nil {
		known[curKey] = make(map[string]Sequence)
	}
	known[curKey][seqKey] = rv
	return rv
}

func RepSeq(key byte, length int) Sequence {
	rv := make(Sequence, length)
	for i := range rv {
		rv[i] = key
	}
	return rv
}

func MovePoint(p *Point, dir byte) *Point {
	if p == nil || dir == A {
		return p
	}
	d := DirDelta[dir]
	if d == nil {
		panic(fmt.Errorf("unknown direction (%d)", dir))
	}
	return NewPoint(p.X+d.X, p.Y+d.Y)
}

type Sequence []byte

func (s Sequence) String() string {
	if len(s) == 0 {
		return "{empty}"
	}
	return string(s)
}

func (p *Point) Follow(seq Sequence) *Point {
	if p == nil || len(seq) == 0 {
		return p
	}
	rv := p
	for _, step := range seq {
		rv = MovePoint(rv, step)
	}
	return rv
}

type Code struct {
	Keys   []byte
	Digits []int
	Value  int
}

var isCodeRx = regexp.MustCompile(`^\d\d\dA$`)

func ParseCode(line string) (*Code, error) {
	if !isCodeRx.MatchString(line) {
		return nil, errors.New("invalid code format")
	}
	digits := line[:3]
	val, err := strconv.Atoi(digits)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q into value: %w", digits, err)
	}
	vals, err := SplitParseIntsD(digits, "")
	if err != nil {
		return nil, fmt.Errorf("could not parse %q into individual digits: %w", digits, err)
	}
	return &Code{Keys: []byte(line), Digits: vals, Value: val}, nil
}

func (c *Code) String() string {
	if c == nil {
		return NilStr
	}
	return string(c.Keys)
}

type Input struct {
	Codes []*Code
}

func (i Input) String() string {
	// StringNumberJoin(slice, startAt, sep) string
	// StringNumberJoinFunc(slice, stringer, startAt, sep) string
	// SliceToStrings(slice) []string
	// AddLineNumbers(lines, startAt) []string
	// MapSlice(slice, mapper) slice  or  MapPSlice  or  MapSliceP
	// CreateIndexedGridString(grid, color, highlight) string  or  CreateIndexedGridStringBz  or  CreateIndexedGridStringNums
	// CreateIndexedGridStringFunc(grid, converter, color, highlight)
	return fmt.Sprintf("(%d): %s", len(i.Codes), StringNumberJoin(i.Codes, 1, ", "))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{Codes: make([]*Code, len(lines))}
	var err error
	for i, line := range lines {
		rv.Codes[i], err = ParseCode(line)
		if err != nil {
			return &rv, fmt.Errorf("invalid input line %d %q: %w", i+1, line, err)
		}
	}
	return &rv, nil
}

func GetB[E any, P XY](grid [][]E, p P) (E, bool) {
	x, y := p.GetXY()
	if y >= 0 && y < len(grid) && x >= 0 && x < len(grid[y]) {
		return grid[y][x], true
	}
	var rv E
	return rv, false
}

func Get[E any, P XY](grid [][]E, p P) E {
	rv, _ := GetB(grid, p)
	return rv
}

func IsIn[E any, P XY](grid [][]E, p P) bool {
	_, ok := GetB(grid, p)
	return ok
}

func (p *Point) Sub(p2 *Point) *Point {
	if p == nil {
		return nil
	}
	if p2 == nil {
		return p
	}
	return NewPoint(p.X-p2.X, p.Y-p2.Y)
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

func (p *Point) Equals(p2 *Point) bool {
	if p == p2 {
		return true
	}
	if p == nil || p2 == nil {
		return false
	}
	return p.X == p2.X && p.Y == p2.Y
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
