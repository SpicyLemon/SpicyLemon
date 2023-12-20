package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_COUNT = -1

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	// Debugf("Parsed Input:\n%s", input)
	modules := NewModuleMap(input.Modules, Start)
	if modules.Map["rx"] == nil {
		Debugf("Parsed Input:\n%s", input)
		return "", fmt.Errorf("no rx module exists in input")
	}
	switch {
	case debug:
		Debugf("There are %d modules:\n%s", modules.Length(), modules)
	case params.Verbose:
		Stdoutf("There are %d modules.", modules.Length())
	}
	if len(params.Custom) == 1 && params.Custom[0] == "mermaid" {
		OutputMermaid(modules)
	}
	answer := CountButtonPresses(modules, params)
	return fmt.Sprintf("%d", answer), nil
}

func CountButtonPresses(modules *ModuleMap, params *Params) int {
	defer FuncEnding(FuncStarting())
	starts := modules.Starter().SendTo
	chains := make(ChainMaps, len(starts))
	for i, start := range starts {
		chains[i] = NewChainMap(modules, start)
	}

	allFound := true
	prevPresses := 0
	for _, chain := range chains {
		FindChainCycle(chain, params, prevPresses)
		if chain.Cycle == 0 {
			allFound = false
			break
		}
		prevPresses += chain.Cycle
	}

	if allFound && params.Verbose {
		Stdoutf("Chain Maps: %s", chains)
	}

	rv := 1
	for _, chain := range chains {
		rv *= chain.Cycle
	}
	return rv
}

func FindChainCycle(chain *ChainMap, params *Params, prevPresses int) {
	defer FuncEnding(FuncStarting())
	presses := 0
	var prev *ChainMap
	if debug {
		prev = chain.Copy()
	}
	for {
		presses++
		pulsesMap := HitButton(chain)
		if !debug && params.Verbose && presses%1_000_000 == 0 {
			Stderrf("%s: Button presses: %d million + %d", chain.SimpleString(), presses/1_000_000, prevPresses)
		}
		if debug {
			ss := chain.SimpleString()
			Stderrf("%s: After press %d:\n%s", ss, presses, GetChanges(prev, chain))
			Stderrf("%s: Pulses sent from end (%d):\n%s", ss, len(pulsesMap[chain.End]), pulsesMap[chain.End])
			prev = chain.Copy()
		}

		for _, p := range pulsesMap[chain.End] {
			if p.Pulse {
				Stdoutf("%s: %d: High pulse sent from end: %s", chain.SimpleString(), presses, p)
				chain.Cycle = presses
			}
		}

		if chain.IsZeroState() {
			Debugf("%s: Chain back to zero state: %s", chain.SimpleString(), chain)
			chain.Cycle = presses
		}

		if chain.Cycle != 0 {
			break
		}

		if 0 < params.Count && params.Count <= presses {
			Stdoutf("%s: Cutoff of %d presses reached.", chain.SimpleString(), presses+prevPresses)
			break
		}
	}
}

type ModuleMap struct {
	Start string
	Map   map[string]*Module
	Order []string
}

func NewModuleMap(modules []*Module, start string) *ModuleMap {
	rv := &ModuleMap{Start: start}
	rv.Map = WireModules(modules)
	rv.Order = MapSlice(SortModules(rv.Map, start), (*Module).GetName)
	return rv
}

func (m ModuleMap) Copy() *ModuleMap {
	return &ModuleMap{
		Start: m.Start,
		Map:   CopyModules(m.Map),
		Order: m.Order,
	}
}

func (m ModuleMap) String() string {
	return StringNumberJoinFunc(m.GetModules(), (*Module).SimpleString, 1, "\n")
}

func (m ModuleMap) GetModules() []*Module {
	rv := make([]*Module, len(m.Order))
	for i, module := range m.Order {
		rv[i] = m.Map[module]
	}
	return rv
}

func (m ModuleMap) Length() int {
	return len(m.Order)
}

func (m ModuleMap) Starter() *Module {
	return m.Map[m.Start]
}

func (m ModuleMap) GetMap() map[string]*Module {
	return m.Map
}

func (m ModuleMap) Reset() {
	for _, v := range m.Map {
		v.Reset()
	}
}

type ChainMaps []*ChainMap

func (c ChainMaps) String() string {
	parts := make([]string, len(c))
	for i, chain := range c {
		parts[i] = PrefixLines(fmt.Sprintf("[%d]", i+1), chain.String())
	}
	return fmt.Sprintf("Chains (%d):\n%s", len(c), strings.Join(parts, "\n"))
}

type ChainMap struct {
	Start string
	End   string
	Map   map[string]*Module
	Order []string
	Cycle int
}

func NewChainMap(modules *ModuleMap, start string) *ChainMap {
	rv := &ChainMap{
		Start: start,
		Map:   make(map[string]*Module),
	}
	chain := GetChain(modules.Map, start)
	rv.Order = make([]string, len(chain))
	for i, m := range chain {
		rv.Map[m.Name] = m
		rv.Order[i] = m.Name
	}
	rv.End = rv.Order[len(rv.Order)-1]
	return rv
}

func (c ChainMap) Copy() *ChainMap {
	return &ChainMap{
		Start: c.Start,
		End:   c.End,
		Map:   CopyModules(c.Map),
		Order: c.Order,
		Cycle: c.Cycle,
	}
}

func (c ChainMap) String() string {
	return fmt.Sprintf("%s->%s = %d\n%s", c.Start, c.End, c.Cycle,
		StringNumberJoinFunc(c.GetModules(), (*Module).SimpleString, 1, "\n"))
}

func (c ChainMap) SimpleString() string {
	return fmt.Sprintf("%s->%s", c.Start, c.End)
}

func (c ChainMap) GetModules() []*Module {
	rv := make([]*Module, len(c.Order))
	for i, module := range c.Order {
		rv[i] = c.Map[module]
	}
	return rv
}

func (c ChainMap) Starter() *Module {
	return c.Map[c.Start]
}

func (c ChainMap) GetMap() map[string]*Module {
	return c.Map
}

func (c ChainMap) IsZeroState() bool {
	// Debugf("Checking for zero state: %s", c)
	return IsZeroState(c.Map)
}

func (c ChainMap) Reset() {
	for _, v := range c.Map {
		v.Reset()
	}
}

func CopyModules(orig map[string]*Module) map[string]*Module {
	if orig == nil {
		return nil
	}
	rv := make(map[string]*Module, len(orig))
	for k, m := range orig {
		rv[k] = m.Copy()
	}
	return rv
}

func SortModules(moduleMap map[string]*Module, start string) []*Module {
	if start != Start {
		return GetChain(moduleMap, start)
	}
	rv := make([]*Module, 1, len(moduleMap))
	rv[0] = moduleMap[start]
	chains := rv[0].SendTo
	for _, head := range chains {
		chain := GetChain(moduleMap, head)
		rv = append(rv, chain...)
	}
	for _, final := range []string{"lg", "rx"} {
		if m := moduleMap[final]; m != nil {
			rv = append(rv, m)
		}
	}
	return rv
}

func GetChain(moduleMap map[string]*Module, start string) []*Module {
	done := make(map[string]bool, 14)
	rv := make([]*Module, 0, 14)
	cjs := make([]*Module, 0, 2)
	queue := []*Module{moduleMap[start]}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if done[cur.Name] {
			continue
		}
		if cur.Type == Conjunction {
			cjs = append(cjs, cur)
		} else {
			rv = append(rv, cur)
		}
		for _, dest := range cur.Dests {
			if done[dest.Name] || dest.Name == "lg" {
				continue
			}
			queue = append(queue, dest)
		}
		done[cur.Name] = true
	}

	if len(cjs) != 2 {
		panic(fmt.Errorf("%d conjunctions found: %s", len(cjs), cjs))
	}
	if len(cjs[0].SendTo) < len(cjs[1].SendTo) {
		cjs[0], cjs[1] = cjs[1], cjs[0]
	}

	return append(rv, cjs...)
}

type HasModules interface {
	GetModules() []*Module
}

func GetChanges(mm1, mm2 HasModules) string {
	var diffs []string
	m1 := mm1.GetModules()
	m2 := mm2.GetModules()
	for i := range m1 {
		if diff := m1[i].Diff(m2[i]); len(diff) > 0 {
			diffs = append(diffs, fmt.Sprintf("%2d: %s", i, diff))
		}
	}
	return strings.Join(diffs, "\n")
}

func OutputMermaid(modules *ModuleMap) {
	lines := make([]string, 0, 110)
	classLines := make([]string, 1, 15)
	classLines[0] = fmt.Sprintf("class %s starter", modules.Start)
	if modules.Start == Start {
		for _, dest := range modules.Starter().Dests {
			classLines = append(classLines, fmt.Sprintf("class %s starter", dest.FullName()))
		}
	}
	for _, mName := range modules.Order {
		cur := modules.Map[mName]
		for i, dest := range cur.Dests {
			lines = append(lines, fmt.Sprintf("%s -->|%d|%s", cur.FullName(), i+1, dest.FullName()))
		}
		if cur.Type == Conjunction {
			classLines = append(classLines, fmt.Sprintf("class %s conjunction", cur.FullName()))
		}
	}
	classLines = append(classLines,
		"classDef starter fill:#33FF99,stroke:#000000,stroke-width:2px",
		"classDef conjunction fill:#FFCC99,stroke:#000000,stroke-width:2px",
	)
	parts := []string{
		"%%{ init: { 'flowchart': { 'curve': 'linear'} } }%%",
		"flowchart TD",
		PrefixLines("    ", strings.Join(lines, "\n")),
		PrefixLines("    ", strings.Join(classLines, "\n")),
	}

	Stdoutf("Mermaid Syntax:\n%s", strings.Join(parts, "\n"))
}

type IsModuleMapper interface {
	Starter() *Module
	GetMap() map[string]*Module
}

func HitButton(modules IsModuleMapper) map[string]Pulses {
	defer FuncEnding(FuncStarting())
	rv := make(map[string]Pulses)
	var queue Pulses
	queue = append(queue, NewPulse("button", Low, modules.Starter().Name))
	moduleMap := modules.GetMap()
	n := 0
	for len(queue) > 0 {
		pulse := queue[0]
		queue = queue[1:]
		n++
		pulse.N = n
		nexts := ReceivePulse(moduleMap, pulse)
		queue = append(queue, nexts...)
		rv[pulse.Source] = append(rv[pulse.Source], pulse)
	}
	return rv
}

func IsZeroState(moduleMap map[string]*Module) bool {
	for _, m := range moduleMap {
		if !m.IsZeroState() {
			return false
		}
	}
	return true
}

func CountHighsLows(pulses Pulses) (int, int) {
	var highs, lows int
	for _, p := range pulses {
		if p.Pulse {
			highs++
		} else {
			lows++
		}
	}
	return highs, lows
}

func SumInts(ints []int) int {
	rv := 0
	for _, i := range ints {
		rv += i
	}
	return rv
}

const (
	Broadcaster = byte('b')
	FlipFlop    = byte('%')
	Conjunction = byte('&')
	End         = byte(' ')
	Low         = false
	High        = true
	Start       = "broadcaster"
)

func ReceivePulse(modules map[string]*Module, pulse *Pulse) Pulses {
	receiver := modules[pulse.Dest]
	if receiver == nil {
		return nil
	}
	var output bool
	switch receiver.Type {
	case Broadcaster:
		output = pulse.Pulse
	case FlipFlop:
		if pulse.Pulse {
			return nil
		}
		receiver.On = !receiver.On
		output = receiver.On
	case Conjunction:
		receiver.Memory[pulse.Source] = pulse.Pulse
		output = receiver.NoLowsInMemory()
	}
	rv := make([]*Pulse, len(receiver.SendTo))
	for i, dest := range receiver.SendTo {
		rv[i] = NewPulse(receiver.Name, output, dest)
	}
	return rv
}

type Pulses []*Pulse

func (p Pulses) String() string {
	return fmt.Sprintf("Pulses (%d):\n%s", len(p), StringNumberJoin(p, 1, "\n"))
}

type Pulse struct {
	N      int
	Source string
	Pulse  bool
	Dest   string
}

func NewPulse(source string, pulse bool, dest string) *Pulse {
	return &Pulse{Source: source, Pulse: pulse, Dest: dest}
}

func (p Pulse) String() string {
	arrow := "low"
	if p.Pulse {
		arrow = "high"
	}
	return fmt.Sprintf("[%d]: %s -%s-> %s", p.N, p.Source, arrow, p.Dest)
}

func WireModules(modules []*Module) map[string]*Module {
	rv := make(map[string]*Module)
	dests := make(map[string]bool)
	alreadyLinked := false
	for _, module := range modules {
		rv[module.Name] = module
		for _, dest := range module.SendTo {
			dests[dest] = true
		}
		alreadyLinked = alreadyLinked || len(module.Dests) > 0 || len(module.Sources) > 0
	}

	for dest := range dests {
		if rv[dest] == nil {
			rv[dest] = EndModule(dest)
		}
	}

	// If they're already linked, we're done.
	for _, module := range rv {
		if len(module.Dests) > 0 || len(module.Sources) > 0 {
			return rv
		}
	}

	for _, source := range modules {
		for _, dest := range source.SendTo {
			ConnectModules(source, rv[dest])
		}
	}

	return rv
}

func ConnectModules(source, dest *Module) {
	source.Dests = append(source.Dests, dest)
	dest.Sources = append(dest.Sources, source)
	if dest.Type == Conjunction {
		dest.Memory[source.Name] = false
	}
}

type Module struct {
	Type    byte
	Name    string
	On      bool
	Memory  map[string]bool
	SendTo  []string
	Dests   []*Module
	Sources []*Module
}

func CopyMap[M ~map[K]V, K comparable, V any](toCopy M) M {
	if toCopy == nil {
		return nil
	}
	rv := make(M, len(toCopy))
	for k, v := range toCopy {
		rv[k] = v
	}
	return rv
}

func CopySlice[S ~[]E, E any](toCopy S) S {
	if toCopy == nil {
		return nil
	}
	rv := make(S, len(toCopy))
	copy(rv, toCopy)
	return rv
}

var ModuleRx = regexp.MustCompile(`^([%&][[:alpha:]]+|broadcaster) -> ([[:alpha:], ]+)$`)

func ParseModule(line string) (*Module, error) {
	parts := ModuleRx.FindStringSubmatch(line)
	if len(parts) != 3 {
		return nil, fmt.Errorf("unknown line format %q", line)
	}

	rv := &Module{
		SendTo: strings.Split(parts[2], ", "),
	}
	if parts[1] == Start {
		rv.Type = Broadcaster
		rv.Name = parts[1]
	} else {
		rv.Type = parts[1][0]
		rv.Name = parts[1][1:]
	}
	if rv.Type == Conjunction {
		rv.Memory = make(map[string]bool)
	}

	return rv, nil
}

func EndModule(name string) *Module {
	return &Module{Type: End, Name: name}
}

func (m Module) Copy() *Module {
	return &Module{
		Type:    m.Type,
		Name:    m.Name,
		On:      m.On,
		Memory:  CopyMap(m.Memory),
		SendTo:  CopySlice(m.SendTo),
		Dests:   CopySlice(m.Dests),
		Sources: CopySlice(m.Sources),
	}
}

func (m Module) String() string {
	parts := make([]string, 1, 2)
	parts[0] = fmt.Sprintf("%s -> %s", m.FullName(), m.SendTo)
	if m.Type == Conjunction {
		parts = append(parts, m.MemoryString())
	}
	if m.Type == FlipFlop {
		parts = append(parts, m.OnString())
	}
	return strings.Join(parts, " ")
}

func (m Module) GetName() string {
	return m.Name
}

func (m Module) FullName() string {
	if m.Type == Broadcaster {
		return m.Name
	}
	return fmt.Sprintf("%c%s", m.Type, m.Name)
}

func (m Module) OnString() string {
	if m.On {
		return "<O>"
	}
	return ">X<"
}

func (m Module) MemoryString() string {
	froms := make([]string, 0, len(m.Memory))
	for k := range m.Memory {
		froms = append(froms, k)
	}
	slices.Sort(froms)
	for i, from := range froms {
		froms[i] += "=" + HLString(m.Memory[from])
	}
	return fmt.Sprintf("{%s}", strings.Join(froms, ","))
}

func HLString(high bool) string {
	if high {
		return "H"
	}
	return "L"
}

func (m Module) SimpleString() string {
	rv := m.Name
	if m.Type == Conjunction {
		rv += ": " + m.MemoryString()
	}
	if m.Type == FlipFlop {
		rv += " = " + m.OnString()
	}
	return rv
}

func (m Module) NoLowsInMemory() bool {
	for _, v := range m.Memory {
		if !v {
			return High
		}
	}
	return Low
}

func (m Module) IsZeroState() bool {
	if m.On {
		return false
	}
	if len(m.Memory) > 0 {
		for _, v := range m.Memory {
			if v {
				return false
			}
		}
	}
	return true
}

func (m Module) Diff(m2 *Module) string {
	if m.Type != m2.Type || m.Name != m2.Name {
		return fmt.Sprintf("%s -> %s", RedString(m.String()), GreenString(m2.String()))
	}
	if m.On != m2.On {
		if m.On {
			return RedString(m2.SimpleString())
		}
		return GreenString(m2.SimpleString())
	}
	if memDiff := m.MemoryDiff(m2); len(memDiff) > 0 {
		return fmt.Sprintf("%s: {%s}", m.Name, memDiff)
	}
	return ""
}

func (m Module) MemoryDiff(m2 *Module) string {
	if m.Memory == nil || m2.Memory == nil {
		return ""
	}
	keys := make([]string, 0, len(m.Memory))
	for k := range m.Memory {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	var parts []string
	haveDiff := false
	for _, k := range keys {
		part := k + "=" + HLString(m2.Memory[k])
		switch {
		case m.Memory[k] == m2.Memory[k]:
			parts = append(parts, part)
		case m.Memory[k]:
			parts = append(parts, RedString(part))
			haveDiff = true
		default:
			parts = append(parts, GreenString(part))
			haveDiff = true
		}
	}
	if haveDiff {
		return strings.Join(parts, ",")
	}
	return ""
}

func RedString(str string) string {
	return "\033[31m" + str + "\033[0m"
}

func GreenString(str string) string {
	return "\033[32m" + str + "\033[0m"
}

func (m *Module) Reset() {
	m.On = false
	for k := range m.Memory {
		m.Memory[k] = Low
	}
}

type Input struct {
	Modules []*Module
}

func (i Input) String() string {
	return fmt.Sprintf("Modules (%d):\n%s", len(i.Modules), StringNumberJoin(i.Modules, 1, "\n"))
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{Modules: make([]*Module, len(lines))}
	var err error
	for i, line := range lines {
		rv.Modules[i], err = ParseModule(line)
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
	for _, entry := range strings.Split(s, sep) {
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

// StringNumberJoin maps the slice to strings, numbers them and joins them.
func StringNumberJoin[S ~[]E, E Stringer](slice S, startAt int, sep string) string {
	return strings.Join(AddLineNumbers(SliceToStrings(slice), startAt), sep)
}

// StringNumberJoin maps the slice to strings, numbers them and joins them.
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
func (c Params) String() string {
	defer FuncEnding(FuncStarting())
	nameFmt := "%10s: "
	lines := []string{
		fmt.Sprintf(nameFmt+"%t", "Debug", debug),
		fmt.Sprintf(nameFmt+"%t", "Verbose", c.Verbose),
		fmt.Sprintf(nameFmt+"%d", "Errors", len(c.Errors)),
		fmt.Sprintf(nameFmt+"%d", "Count", c.Count),
		fmt.Sprintf(nameFmt+"%s", "Input File", c.InputFile),
		fmt.Sprintf(nameFmt+"%d lines", "Input", len(c.Input)),
		fmt.Sprintf(nameFmt+"%d lines", "Custom", len(c.Custom)),
	}
	if len(c.Errors) > 0 {
		lines = append(lines, fmt.Sprintf("Errors (%d):", len(c.Errors)))
		errors := make([]string, len(c.Errors))
		for i, err := range c.Errors {
			errors[i] = err.Error()
		}
		lines = append(lines, AddLineNumbers(errors, 1)...)
	}
	if len(c.Input) > 0 {
		lines = append(lines, fmt.Sprintf("Input (%d):", len(c.Input)))
		lines = append(lines, AddLineNumbers(c.Input, 0)...)
	}
	if len(c.Custom) > 0 {
		lines = append(lines, fmt.Sprintf("Custom Input (%d):", len(c.Custom)))
		lines = append(lines, AddLineNumbers(c.Custom, 0)...)
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
func (c *Params) AppendError(err error) {
	if err != nil {
		c.Errors = append(c.Errors, err)
	}
}

// HasError returns true if this Params has one or more errors.
func (c Params) HasError() bool {
	return len(c.Errors) != 0
}

// Error flattens the Errors slice into a single string.
// It also makes the Params struct satisfy the error interface.
func (c Params) GetError() error {
	switch len(c.Errors) {
	case 0:
		return nil
	case 1:
		return c.Errors[0]
	default:
		errs := make([]error, 1, 1+len(c.Errors))
		errs[0] = fmt.Errorf("Found %d errors:", len(c.Errors)) //nolint:stylecheck,revive // punct okay here.
		for i, err := range c.Errors {
			errs = append(errs, fmt.Errorf("  %d: %w", i+1, err))
		}
		return errors.Join(errs...)
	}
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
