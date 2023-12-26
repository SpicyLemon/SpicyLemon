package main

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
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
	answer := FindSolution(input.HailStones, params)
	return fmt.Sprintf("%d", answer), nil
}

func FindSolution(stones HailStones, params *Params) *big.Int {
	defer FuncEnding(FuncStarting())

	var xOpts, yOpts, zOpts []int
	for i := 0; i < len(stones)-1; i++ {
		for j := i + 1; j < len(stones); j++ {
			a, b := stones[i], stones[j]
			xOpts = FindOpts(xOpts, b.X-a.X, a.DX, b.DX)
			yOpts = FindOpts(yOpts, b.Y-a.Y, a.DY, b.DY)
			zOpts = FindOpts(zOpts, b.Z-a.Z, a.DZ, b.DZ)
		}
	}

	params.Verbosef("Options:\nDX: %v\nDY: %v\nDZ: %v", xOpts, yOpts, zOpts)

	if len(xOpts) != 1 || len(yOpts) != 1 || len(zOpts) != 1 {
		Stderrf("No solution found.")
		return nil
	}

	stone := CalcStone(stones, xOpts[0], yOpts[0], zOpts[0])

	if !IsSolution(stone, stones) {
		Debugf("Fuck.")
	}

	rv := &big.Int{}
	rv.Add(rv, NewBigInt(stone.X))
	rv.Add(rv, NewBigInt(stone.Y))
	rv.Add(rv, NewBigInt(stone.Z))
	return rv
}

func CalcStone(stones HailStones, dx, dy, dz int) *HailStone {
	stone1, stone2 := stones[0], stones[1]
	m1, b1 := GetDMB(stone1, dx, dy)
	m2, b2 := GetDMB(stone2, dx, dy)

	xTop := new(big.Float).Sub(b2, b1)
	xBot := new(big.Float).Sub(m1, m2)
	xf := new(big.Float).Quo(xTop, xBot)
	x := BigFloatToInt(xf)
	xf = NewBigFloat(x)
	Debugf("X: %d <= %s", x, xf.Text('f', 5))

	yf := new(big.Float).Mul(m1, xf)
	yf.Add(yf, b1)
	y := BigFloatToInt(yf)
	yf = NewBigFloat(y)
	Debugf("Y: %d <= %s", y, yf.Text('f', 5))

	tTop := new(big.Float).Sub(xf, NewBigFloat(stone1.X))
	tBot := new(big.Float).Sub(NewBigFloat(stone1.DX), NewBigFloat(dx))
	tf := new(big.Float).Quo(tTop, tBot)
	t := BigFloatToInt(tf)
	tf = NewBigFloat(t)
	Debugf("T: %d <= %s", t, tf.Text('f', 5))

	zf := new(big.Float).Sub(NewBigFloat(stone1.DZ), NewBigFloat(dz))
	zf.Mul(zf, tf)
	zf.Add(zf, NewBigFloat(stone1.Z))
	z := BigFloatToInt(zf)
	zf = NewBigFloat(z)
	Debugf("Z: %d <= %s", z, zf.Text('f', 5))

	Debugf("Solution:\n X: %s\nDX: %d\n Y: %s\nDY: %d\n Z: %s\nDZ: %d",
		xf.Text('f', 5), dx, yf.Text('f', 5), dy, zf.Text('f', 5), dz)
	if !xf.IsInt() {
		Stdoutf("X: %s is not an integer.", xf.Text('f', 5))
	}
	if !yf.IsInt() {
		Stdoutf("Y: %s is not an integer.", yf.Text('f', 5))
	}
	if !zf.IsInt() {
		Stdoutf("Z: %s is not an integer.", zf.Text('f', 5))
	}

	rv := &HailStone{
		X:  x,
		Y:  y,
		Z:  z,
		DX: dx,
		DY: dy,
		DZ: dz,
	}

	Debugf("Rock: %s", rv)

	return rv
}

func GetDMB(stone *HailStone, dx, dy int) (*big.Float, *big.Float) {
	mTop := new(big.Float).Sub(NewBigFloat(stone.DY), NewBigFloat(dy))
	mBot := new(big.Float).Sub(NewBigFloat(stone.DX), NewBigFloat(dx))
	m := new(big.Float).Quo(mTop, mBot)

	b := new(big.Float).Mul(m, NewBigFloat(stone.X))
	b.Sub(NewBigFloat(stone.Y), b)
	return m, b
}

func CalcXYZFail(stones HailStones, sdx, sdy, sdz int) *HailStone {
	dx := NewBigInt(sdx)
	dy := NewBigInt(sdy)
	stone1, stone2 := stones[0], stones[1]
	dx1, dy1, dz1 := stone1.GetBigDXYZ()
	dx2, dy2 := stone2.GetBigDXY()
	x1, y1, z1 := stone1.GetBigXYZ()
	x2, y2 := stone2.GetBigXY()

	ddx1 := BigSub(dx1, dx)
	ddx2 := BigSub(dx2, dx)
	ddy1 := BigSub(dy1, dy)
	ddy2 := BigSub(dy2, dy)
	yd21 := BigSub(y2, y1)
	xTopL := BigMul(yd21, ddx1)
	xTopL.Mul(xTopL, ddx2)
	xTopM := BigMul(x1, ddy1)
	xTopM.Mul(xTopM, ddx2)
	xTopR := BigMul(x2, ddy2)
	xTopR.Mul(xTopR, ddx1)
	xTop := BigAdd(xTopL, xTopM)
	xTop.Sub(xTop, xTopR)

	xBotL := BigMul(ddy1, ddx2)
	xBotR := BigMul(ddy2, ddx1)
	xBot := BigSub(xBotL, xBotR)
	x, xRem := BigQuoRem(xTop, xBot)
	if !BigIsZero(xRem) {
		Stdoutf("X: %s / %s has a non-zero remainder of %s", xTop, xBot, xRem)
		return nil
	}
	Debugf("X: %s", x)

	t := BigIntToBigFloat(x)
	t.Sub(t, BigIntToBigFloat(x1))
	t.Quo(t, BigIntToBigFloat(dx1))
	Debugf("%s has x = %s at time %s", stone1, x, t)

	y := t.Mul(t, BigIntToBigFloat(dy1))
	y.Add(y, BigIntToBigFloat(y1))
	if !y.IsInt() {
		Stdoutf("Y: %s * %s + %s = %s which is not an int.", t.String(), dy1, y1, y.String())
		return nil
	}
	z := t.Mul(t, BigIntToBigFloat(dz1))
	z.Add(z, BigIntToBigFloat(z1))
	if !z.IsInt() {
		Stdoutf("Z: %s * %s + %s = %s which is not an int.", t.String(), dz1, z1, z.String())
		return nil
	}

	rv := &HailStone{
		X:  int(x.Int64()),
		Y:  int(BigFloatToBigInt(y).Int64()),
		Z:  int(BigFloatToBigInt(z).Int64()),
		DX: sdx,
		DY: sdy,
		DZ: sdz,
	}

	Debugf("Solution: %s", rv)
	return rv
}

func FindOpts(opts []int, p, ad, bd int) []int {
	if ad != bd {
		return opts
	}
	newOpts := GetMatchingD(p, ad)
	if len(opts) == 0 {
		return newOpts
	}
	return Intersect(opts, newOpts)
}

func GetMatchingD(p, d int) []int {
	var rv []int
	for v := -1000; v <= 1000; v++ {
		if v != d && p%(v-d) == 0 {
			rv = append(rv, v)
		}
	}
	return rv
}

func Intersect[S ~[]E, E comparable](a, b S) S {
	var rv S
	for _, v := range a {
		if slices.Contains(b, v) {
			rv = append(rv, v)
		}
	}
	return rv
}

func IsSolution(guess *HailStone, stones HailStones) bool {
	for i, stone := range stones {
		if p := guess.GetIntersection(stone); p == nil {
			Debugf("%d: %s does not intersect %s", i, stone, guess)
			return false
		}
	}
	return true
}

func NewBigFloat[I Signed](val I) *big.Float {
	return new(big.Float).SetInt(NewBigInt(val))
}

var (
	RounderP = big.NewFloat(0.5)
	RounderN = big.NewFloat(-0.5)
)

func BigFloatToInt(bf *big.Float) int {
	rvf := new(big.Float)
	if bf.Sign() < 0 {
		rvf.Add(bf, RounderN)
	} else {
		rvf.Add(bf, RounderP)
	}
	rv, _ := rvf.Int64()
	return int(rv)
}

func BigFloatToBigInt(bf *big.Float) *big.Int {
	rv, _ := bf.Int(nil)
	return rv
}

func BigIntToBigFloat(bi *big.Int) *big.Float {
	return new(big.Float).SetInt(bi)
}

func NewBigInt[I Signed](val I) *big.Int {
	return big.NewInt(int64(val))
}

func BigSum(vals ...int) *big.Int {
	rv := &big.Int{}
	for _, val := range vals {
		rv.Add(rv, NewBigInt(val))
	}
	return rv
}

func BigAdd(a, b *big.Int) *big.Int {
	rv := &big.Int{}
	return rv.Add(a, b)
}

func BigSub(a, b *big.Int) *big.Int {
	rv := &big.Int{}
	return rv.Sub(a, b)
}

func BigMul(a, b *big.Int) *big.Int {
	rv := &big.Int{}
	return rv.Mul(a, b)
}

func BigQuoRem(a, b *big.Int) (*big.Int, *big.Int) {
	rv := &big.Int{}
	return rv.QuoRem(a, b, &big.Int{})
}

var Zero = big.NewInt(0)

func BigIsZero(a *big.Int) bool {
	return Zero.Cmp(a) == 0
}

func ShiftStones(stones HailStones, limit *MinMax[int]) (HailStones, *MinMax[int]) {
	rv := make(HailStones, len(stones))
	for i, stone := range stones {
		rv[i] = &HailStone{
			X:  stone.X - limit.Min,
			Y:  stone.Y - limit.Min,
			Z:  stone.Z - limit.Min,
			DX: stone.DX,
			DY: stone.DY,
			DZ: stone.DZ,
		}
	}
	return rv, NewMinMax(0, limit.Max-limit.Min)
}

type Float interface {
	~float32 | ~float64
}

type Integer interface {
	Signed | Unsigned
}

type Ordered interface {
	Integer | Float | ~string
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func OrderNums[E Ordered](a, b E) (E, E) {
	if a <= b {
		return a, b
	}
	return b, a
}

type MinMax[E Ordered] struct {
	Min E
	Max E
}

func NewMinMax[E Ordered](min, max E) *MinMax[E] {
	min, max = OrderNums(min, max)
	return &MinMax[E]{Min: min, Max: max}
}

func (m MinMax[E]) String() string {
	_, isFloat := any(m.Min).(float64)
	if !isFloat {
		_, isFloat = any(m.Min).(float32)
	}
	if isFloat {
		return fmt.Sprintf("[%.4f to %.4f]", m.Min, m.Max) //nolint:govet // Definitely a float here.
	}
	return fmt.Sprintf("[%d to %d]", m.Min, m.Max) //nolint:govet // Definitely an int here.
}

func (m MinMax[E]) Contains(n E) bool {
	return m.Min <= n && n <= m.Max
}

func (m MinMax[E]) IsValid() bool {
	return m.Min <= m.Max
}

func (m *MinMax[E]) Include(val E) {
	if val < m.Min {
		m.Min = val
	}
	if val > m.Max {
		m.Max = val
	}
}

func IntersectMinMax[E Ordered](m1 *MinMax[E], m2 *MinMax[E]) *MinMax[E] {
	_, min := OrderNums(m1.Min, m2.Min)
	max, _ := OrderNums(m1.Max, m2.Max)
	// Debugf("Mins:\nm1.Min = %20.4f\nm2.Min = %20.4f\n r.Min = %20.4f", m1.Min, m2.Min, min)
	// Debugf("Maxs:\nm1.Max = %20.4f\nm2.Max = %20.4f\n r.Max = %20.4f", m1.Max, m2.Max, max)
	if min > max {
		// Debugf("No intersection:\n r.Min = %20.4f\n r.Max = %20.4f", min, max)
		return nil
	}
	return NewMinMax(min, max)
}

type HailStone struct {
	X  int
	Y  int
	Z  int
	DX int
	DY int
	DZ int
}

func ParseHailStone(line string) (*HailStone, error) {
	parts := strings.Split(line, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("could not parse hailstone %q", line)
	}
	xyz, err := SplitParseInts(parts[0], ", ")
	if err != nil || len(xyz) != 3 {
		return nil, fmt.Errorf("could not parse position values from %q from hailstone %q: %w",
			parts[0], line, err)
	}
	ds, err := SplitParseInts(parts[1], ", ")
	if err != nil || len(ds) != 3 {
		return nil, fmt.Errorf("could not parse change values from %q from hailstone %q: %w",
			parts[1], line, err)
	}
	return &HailStone{X: xyz[0], Y: xyz[1], Z: xyz[2], DX: ds[0], DY: ds[1], DZ: ds[2]}, nil
}

func (h HailStone) String() string {
	return fmt.Sprintf("(%d,%d,%d)+(%d,%d,%d)", h.X, h.Y, h.Z, h.DX, h.DY, h.DZ)
}

func (h HailStone) GetX() int {
	return h.X
}

func (h HailStone) GetY() int {
	return h.Y
}

func (h HailStone) GetZ() int {
	return h.Z
}

func (h HailStone) GetXY() (int, int) {
	return h.X, h.Y
}

func (h HailStone) GetXZ() (int, int) {
	return h.X, h.Z
}

func (h HailStone) GetYZ() (int, int) {
	return h.Y, h.Z
}

func (h HailStone) GetXYZ() (int, int, int) {
	return h.X, h.Y, h.Z
}

func (h HailStone) GetDX() int {
	return h.DX
}

func (h HailStone) GetDY() int {
	return h.DY
}

func (h HailStone) GetDZ() int {
	return h.DZ
}

func (h HailStone) GetDXY() (int, int) {
	return h.DX, h.DY
}

func (h HailStone) GetDXZ() (int, int) {
	return h.DX, h.DZ
}

func (h HailStone) GetDYZ() (int, int) {
	return h.DY, h.DZ
}

func (h HailStone) GetDXYZ() (int, int, int) {
	return h.DX, h.DY, h.DZ
}

func (h HailStone) GetXDX() (int, int) {
	return h.X, h.DX
}

func (h HailStone) GetYDY() (int, int) {
	return h.Y, h.DY
}

func (h HailStone) GetZDZ() (int, int) {
	return h.Z, h.DZ
}

func (h HailStone) GetTimeAt(x, y float64) float64 {
	if h.DX != 0 {
		return (x - float64(h.X)) / float64(h.DX)
	}
	if h.DY != 0 {
		return (y - float64(h.Y)) / float64(h.DY)
	}
	return 0
}

func (h HailStone) GetXYZAt(t int) (int, int, int) {
	return h.X + h.DX*t, h.Y + h.DY*t, h.Z + h.DZ*t
}

func (h HailStone) Contains(x, y, z int) bool {
	if h.DX == 0 || h.DY == 0 || h.DZ == 0 {
		return false
	}
	fx := (float64(x) - float64(h.X)) / float64(h.DX)
	fy := (float64(y) - float64(h.Y)) / float64(h.DY)
	fz := (float64(z) - float64(h.Z)) / float64(h.DZ)
	return fx == fy && fy == fz
}

type Point struct {
	X, Y, Z int
}

func (h HailStone) GetIntersection(h2 *HailStone) *Point {
	x1, y1 := h.GetBigFXY()
	dx1, dy1 := h.GetBigFDXY()
	x2, y2 := h2.GetBigFXY()
	dx2, dy2 := h2.GetBigFDXY()

	m2 := new(big.Float).Quo(dy2, dx2)
	dx := new(big.Float).Sub(x1, x2)
	dy := new(big.Float).Sub(y1, y2)

	top := new(big.Float).Mul(m2, dx)
	top.Sub(dy, top)

	bot := new(big.Float).Mul(m2, dx1)
	bot.Sub(bot, dy1)

	tf := new(big.Float).Quo(top, bot)
	t := BigFloatToInt(tf)
	x, y, z := h.GetXYZAt(t)
	_, _, z2 := h2.GetXYZAt(t)

	if z != z2 {
		if debug {
			lines := []string{
				"No intersection",
				fmt.Sprintf("L1: %s", h),
				fmt.Sprintf("L2: %s", h2),
				fmt.Sprintf(" T: %d <= %s", t, tf.Text('f', 5)),
				fmt.Sprintf(" X: %d = %d + %d * %d", x, h.X, h.DX, t),
				fmt.Sprintf(" Y: %d = %d + %d * %d", y, h.Y, h.DY, t),
				fmt.Sprintf("Z1: %d = %d + %d * %d", z, h.Z, h.DZ, t),
				fmt.Sprintf("Z2: %d = %d + %d * %d", z2, h2.Z, h2.DZ, t),
			}
			Stderrf("%s", strings.Join(lines, "\n"))
		}
		return nil
	}

	return &Point{X: x, Y: y, Z: z}
}

func (h HailStone) GetBigFXY() (*big.Float, *big.Float) {
	return NewBigFloat(h.X), NewBigFloat(h.Y)
}

func (h HailStone) GetBigFDXY() (*big.Float, *big.Float) {
	return NewBigFloat(h.DX), NewBigFloat(h.DY)
}

func (h HailStone) GetBigX() *big.Int {
	return NewBigInt(h.X)
}

func (h HailStone) GetBigY() *big.Int {
	return NewBigInt(h.Y)
}

func (h HailStone) GetBigZ() *big.Int {
	return NewBigInt(h.Z)
}

func (h HailStone) GetBigXY() (*big.Int, *big.Int) {
	return h.GetBigX(), h.GetBigY()
}

func (h HailStone) GetBigXYZ() (*big.Int, *big.Int, *big.Int) {
	return h.GetBigX(), h.GetBigY(), h.GetBigZ()
}

func (h HailStone) GetBigDX() *big.Int {
	return NewBigInt(h.DX)
}

func (h HailStone) GetBigDY() *big.Int {
	return NewBigInt(h.DY)
}

func (h HailStone) GetBigDZ() *big.Int {
	return NewBigInt(h.DZ)
}

func (h HailStone) GetBigDXY() (*big.Int, *big.Int) {
	return h.GetBigDX(), h.GetBigDY()
}

func (h HailStone) GetBigDXYZ() (*big.Int, *big.Int, *big.Int) {
	return h.GetBigDX(), h.GetBigDY(), h.GetBigDZ()
}

type HailStones []*HailStone

func (h HailStones) String() string {
	return StringNumberJoin(h, 1, "\n")
}

type Input struct {
	HailStones HailStones
}

func (i Input) String() string {
	return fmt.Sprintf("Hail Stones (%d):\n%s", len(i.HailStones),
		StringNumberJoin(i.HailStones, 1, "\n"))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{HailStones: make(HailStones, len(lines))}
	var err error
	for i, line := range lines {
		rv.HailStones[i], err = ParseHailStone(line)
		if err != nil {
			return nil, err
		}
	}
	return &rv, nil
}

// -------------------------------------------------------------------------------------------------
// -------------------------------------  Some generic stuff  --------------------------------------
// -------------------------------------------------------------------------------------------------

const MIN_INT8 = int8(-128)
const MAX_INT8 = int8(127)
const MIN_INT16 = int16(-32_768)
const MAX_INT16 = int16(32_767)
const MIN_INT32 = int32(-2_147_483_648)
const MAX_INT32 = int32(2_147_483_647)
const MIN_INT64 = int64(-9_223_372_036_854_775_808)
const MAX_INT64 = int64(9_223_372_036_854_775_807)
const MIN_INT = -9_223_372_036_854_775_808
const MAX_INT = 9_223_372_036_854_775_807

const MAX_UINT8 = uint8(255)
const MAX_UINT16 = uint16(65_535)
const MAX_UINT32 = uint32(4_294_967_295)
const MAX_UINT64 = uint64(18_446_744_073_709_551_615)
const MAX_UINT = uint(18_446_744_073_709_551_615)

// SplitParseInts splits a string using the given separator and converts each part into an int.
// Uses strings.Split(s, sep) for the splitting and strconv.Atoi to parse it to an int.
// Leading and trailing whitespace on each entry are ignored.
func SplitParseInts(s string, sep string) ([]int, error) {
	rv := []int{}
	for _, entry := range strings.Split(strings.TrimSpace(s), sep) {
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

// StringNumberJoin maps the slice to strings, numbers them, and joins them.
func StringNumberJoin[S ~[]E, E Stringer](slice S, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, E.String), startAt), sep)
}

// StringNumberJoinFunc maps the slice to strings using the provided stringer, numbers them, and joins them.
func StringNumberJoinFunc[S ~[]E, E any](slice S, stringer func(E) string, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(MapSlice(slice, stringer), startAt), sep)
}

// SliceToStrings runs String() on each entry of the provided slice.
func SliceToStrings[S ~[]E, E Stringer](slice S) []string {
	return MapSlice(slice, E.String)
}

// AddLineNumbers adds line numbers to each string.
func AddLineNumbers(lines []string, startAt int) []string {
	if len(lines) == 0 {
		return []string{}
	}
	lineFmt := DigitFormatForMax(len(lines)) + ": %s"
	rv := make([]string, len(lines))
	for i, line := range lines {
		rv[i] = fmt.Sprintf(lineFmt, i+startAt, line)
	}
	return rv
}

// DigitFormatForMax returns a format string of the length of the provided maximum number.
// E.g. DigitFormatForMax(10) returns "%2d".
// DigitFormatForMax(382920) returns "%6d".
func DigitFormatForMax(max int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", max)))
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

// Stringer is an interface for something that can be turned into a string.
type Stringer interface {
	String() string
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

// -------------------------------------------------------------------------------------------------
// --------------------------------  CLI params and input parsing  ---------------------------------
// -------------------------------------------------------------------------------------------------

// Params contains anything that might be provided via command-line arguments.
type Params struct {
	// Verbose is a flag indicating some extra output is desired.
	Verbose bool
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
	defer FuncEnding(FuncStarting())
	nameFmt := "%10s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", debug),
		fmt.Sprintf(nameFmt+"%t", "Verbose", p.Verbose),
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
		case HasPrefixFold(args[i], "--debug"):
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
			rv.Verbose, extraI, err = ParseFlagBool(args[i:])
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
		rv.Verbose = debug
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

// Verbosef outputs to Stderr if the verbose flag was provided. Does nothing otherwise.
func (p Params) Verbosef(format string, a ...interface{}) {
	if p.Verbose {
		StderrAsf(GetFuncName(1), format, a...)
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
	if debug {
		StderrAsf(GetFuncName(1), format, a...)
	} else {
		StdoutAsf(GetFuncName(1), format, a...)
	}
}

// DebugAlwaysAsf is like StderrAsf if the debug flag is set; otherwise it's like StdoutAsf.
func DebugAlwaysAsf(funcName, format string, a ...interface{}) {
	if debug {
		StderrAsf(funcName, format, a...)
	} else {
		StdoutAsf(funcName, format, a...)
	}
}

// -------------------------------------------------------------------------------------------------
// --------------------------------  Primary Program Running Parts  --------------------------------
// -------------------------------------------------------------------------------------------------

// debug is a flag for whether or not debug messages should be displayed.
var debug bool

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
