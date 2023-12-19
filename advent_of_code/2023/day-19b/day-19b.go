package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
	answer := FindAcceptable(input.Workflows)
	return fmt.Sprintf("%d", answer), nil
}

func FindAcceptable(workflows []*Workflow) int {
	defer FuncEnding(FuncStarting())
	wfMap := make(map[string]*Workflow, len(workflows))
	for _, wf := range workflows {
		wfMap[wf.Name] = wf
	}

	toCheck := []*PartRange{InitialPartRange()}

	var accepted, rejected []*PartRange
	for len(toCheck) > 0 {
		var newChecks []*PartRange
		for i, pr := range toCheck {
			Debugf("(%3d)[%2d]: Applying Workflow To: %s", len(toCheck), i, pr)
			nc := ApplyWorkflow(wfMap, pr)
			Debugf("(%3d)[%2d]: %d Results.", len(toCheck), i, len(nc))
			for j, npr := range nc {
				if npr.Next == Reject {
					Debugf("(%3d)[%2d][%d]: REJECTED: %s", len(toCheck), i, j, npr)
					rejected = append(rejected, npr)
					continue
				}
				if npr.Next == Accept {
					Debugf("(%3d)[%2d][%d]: ACCEPTED: %s", len(toCheck), i, j, npr)
					accepted = append(accepted, npr)
					continue
				}
				Debugf("(%3d)[%2d][%d]: not done: %s", len(toCheck), i, j, npr)
				newChecks = append(newChecks, npr)
			}
		}
		toCheck = newChecks
	}

	if debug {
		Stderrf("Accepted (%d):\n%s", len(accepted), StringNumberJoin(accepted, 1, "\n"))
		Stderrf("Rejected (%d):\n%s", len(rejected), StringNumberJoin(rejected, 1, "\n"))
	}

	keepers := SimplifyPartRanges(accepted)
	tossers := SimplifyPartRanges(rejected)

	good := 0
	for i, pr := range keepers {
		nv := pr.Count()
		Debugf("Accepted: %d: %s = %16d", i+1, pr, nv)
		good += nv
	}

	bad := 0
	for i, pr := range tossers {
		nv := pr.Count()
		Debugf("Rejected: %d: %s = %16d", i+1, pr, nv)
		bad += nv
	}

	if debug {
		Stderrf(" Good: %16d", good)
		Stderrf("  Bad: %16d", bad)
		Stderrf("Total: %16d", good+bad)
	}

	return good
}

func SimplifyPartRanges(prs []*PartRange) []*PartRange {
	defer FuncEnding(FuncStarting())

	queue := make([]*PartRange, len(prs), len(prs)*2)
	for i, pr := range prs {
		queue[i] = pr.CopyRanges()
	}
	var keepers []*PartRange
	for len(queue) > 0 {
		pivot := queue[0]
		queue = queue[1:]
		Debugf("Queue: %d", len(queue))
		isKeeper := true
		end := len(queue)
		for i := 0; i < end; i++ {
			other := queue[i]
			newPrs, nextPivot, removeOther := SplitPartRange(pivot, other)
			if nextPivot {
				if len(newPrs) > 0 {
					if debug {
						Stderrf("[%d]: Split Result:\nPivot: %s\nOther: %s\n%s\n%s", i, pivot, other,
							strings.Repeat("-", 60), PrefixLines("    ", StringNumberJoin(newPrs, 1, "\n")))
					}
					queue = append(queue, newPrs...)
				}
				Debugf("[%d]: Moving on to next pivot.", i)
				isKeeper = false
				break
			}
			if removeOther {
				Debugf("[%d]: Removing other: %s", i, other)
				if i+1 < len(queue) {
					copy(queue[i:], queue[i+1:])
				}
				queue[len(queue)-1] = nil
				queue = queue[:len(queue)-1]
				i--
				end--
			}
		}
		if isKeeper {
			Debugf("Keeper: %s", pivot)
			keepers = append(keepers, pivot)
		}
	}

	return keepers
}

// SplitPartRange returns the pivot splits, nextPivot, removeOther.
func SplitPartRange(pivot, other *PartRange) ([]*PartRange, bool, bool) {
	defer FuncEnding(FuncStarting())
	if debug {
		Stderrf("  Pivot: %s", pivot)
		Stderrf("  Other: %s", other)
	}
	overlap := &PartRange{Ranges: make(map[byte]*MinMax, 4)}
	for _, b := range []byte{Cool, Music, Aero, Shiny} {
		nr := CombineMinMaxes(pivot.Ranges[b], other.Ranges[b])
		if nr == nil {
			Debugf("Overlap: None")
			// Move on to next other.
			return nil, false, false
		}
		overlap.Ranges[b] = nr
	}
	Debugf("Overlap: %s", overlap)
	if pivot.Equal(overlap) {
		Debugf("  pivot is completely inside other")
		// Nothing to re-add, move on to next pivot.
		return nil, true, false
	}
	if other.Equal(overlap) {
		Debugf("  other is completely inside pivot")
		// Nothing to re-add, remove this other.
		return nil, false, true
	}

	// Subtract the overlap from pivot by first splitting it up.
	splits := []*PartRange{pivot}
	for _, b := range []byte{Cool, Music, Aero, Shiny} {
		var newSplits []*PartRange
		for _, spr := range splits {
			if spr.Ranges[b].Min < overlap.Ranges[b].Min {
				npr := spr.CopyRanges()
				npr.Ranges[b].Max = overlap.Ranges[b].Min - 1
				newSplits = append(newSplits, npr)
			}
			if overlap.Ranges[b].Max < spr.Ranges[b].Max {
				npr := spr.CopyRanges()
				npr.Ranges[b].Min = overlap.Ranges[b].Max + 1
				newSplits = append(newSplits, npr)
			}
			npr := spr.CopyRanges()
			npr.Ranges[b] = overlap.Ranges[b]
			if b == Shiny && npr.Equal(overlap) {
				continue
			}
			newSplits = append(newSplits, npr)
		}
		splits = newSplits
	}

	return splits, true, false
}

func CombineMinMaxes(m1, m2 *MinMax) *MinMax {
	rv := &MinMax{
		Min: Max(m1.Min, m2.Min),
		Max: Min(m1.Max, m2.Max),
	}
	if !rv.IsValid() {
		return nil
	}
	return rv
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const (
	Cool   = byte('x')
	Music  = byte('m')
	Aero   = byte('a')
	Shiny  = byte('s')
	Accept = "A"
	Reject = "R"
)

func ApplyWorkflow(wfMap map[string]*Workflow, pr *PartRange) []*PartRange {
	var rv []*PartRange
	for _, rule := range wfMap[pr.Next].Rules {
		var npr *PartRange
		npr, pr = ApplyRule(rule, pr)
		if npr != nil {
			rv = append(rv, npr)
		}
	}
	return rv
}

func ApplyRule(rule *Rule, pr *PartRange) (*PartRange, *PartRange) {
	// Rule: Field, Comp, Val, Result
	good, bad := pr.Copy(), pr.Copy()
	good.Next = rule.Result
	switch rule.Comp {
	case '<':
		good.Ranges[rule.Field].Max = rule.Val - 1
		if !good.Ranges[rule.Field].IsValid() {
			good = nil
		}
		bad.Ranges[rule.Field].Min = rule.Val
		if !bad.Ranges[rule.Field].IsValid() {
			bad = nil
		}
	case '>':
		good.Ranges[rule.Field].Min = rule.Val + 1
		if !good.Ranges[rule.Field].IsValid() {
			good = nil
		}
		bad.Ranges[rule.Field].Max = rule.Val
		if !bad.Ranges[rule.Field].IsValid() {
			bad = nil
		}
	case '=':
		bad = nil
		// Do nothing else.
	default:
		panic(fmt.Errorf("unknown rule comp %c in %s", rule.Comp, rule))
	}
	if good != nil && pr.Contains(good) {
		good = nil
	}
	return good, bad
}

type MinMax struct {
	Min int
	Max int
}

func NewMinMax(min, max int) *MinMax {
	return &MinMax{Min: min, Max: max}
}

func (r MinMax) Copy() *MinMax {
	return &MinMax{Min: r.Min, Max: r.Max}
}

func (r MinMax) Count() int {
	return r.Max - r.Min + 1
}

func (r MinMax) IsValid() bool {
	return r.Min <= r.Max
}

func (r MinMax) Equal(r2 *MinMax) bool {
	return r2 != nil && r.Min == r2.Min && r.Max == r2.Max
}

func (r MinMax) String() string {
	return fmt.Sprintf("%4d-%4d", r.Min, r.Max)
}

type PartRange struct {
	Ranges map[byte]*MinMax
	Next   string
	Prev   *PartRange
}

func InitialPartRange() *PartRange {
	return &PartRange{
		Ranges: map[byte]*MinMax{
			Cool:  NewMinMax(1, 4000),
			Music: NewMinMax(1, 4000),
			Aero:  NewMinMax(1, 4000),
			Shiny: NewMinMax(1, 4000),
		},
		Next: "in",
		Prev: nil,
	}
}

func (p PartRange) String() string {
	parts := make([]string, 4)
	for i, b := range []byte{Cool, Music, Aero, Shiny} {
		parts[i] = fmt.Sprintf("%c:%s", b, p.Ranges[b].String())
	}
	return fmt.Sprintf("{%s}: %q = %16d", strings.Join(parts, ", "), p.Next, p.Count())
}

func (p PartRange) Copy() *PartRange {
	rv := p.CopyRanges()
	rv.Next = p.Next
	rv.Prev = &p
	return rv
}

func (p PartRange) CopyRanges() *PartRange {
	rv := &PartRange{
		Ranges: make(map[byte]*MinMax),
	}
	for k, v := range p.Ranges {
		rv.Ranges[k] = v.Copy()
	}
	return rv
}

func (p PartRange) Equal(p2 *PartRange) bool {
	if p.Next != p2.Next {
		return false
	}
	for _, k := range []byte{Cool, Music, Aero, Shiny} {
		if !p.Ranges[k].Equal(p2.Ranges[k]) {
			return false
		}
	}
	return true
}

func (p PartRange) Contains(p2 *PartRange) bool {
	cur := &p
	for cur != nil {
		if cur.Equal(p2) {
			return true
		}
		cur = cur.Prev
	}
	return false
}

func (p PartRange) Count() int {
	rv := 1
	for _, mm := range p.Ranges {
		rv *= mm.Count()
	}
	return rv
}

func RunWorkflows(workflows []*Workflow, parts []*Part) []*Part {
	wfMap := make(map[string]*Workflow, len(workflows))
	for _, wf := range workflows {
		wfMap[wf.Name] = wf
	}

	var accepted, rejected []*Part
	for _, part := range parts {
		if ProcessPart(wfMap, part) {
			accepted = append(accepted, part)
		} else {
			rejected = append(rejected, part)
		}
	}
	Debugf("Accepted (%d):\n%s\nRejected (%d):\n%s",
		len(accepted), StringNumberJoin(accepted, 1, "\n"),
		len(rejected), StringNumberJoin(rejected, 1, "\n"),
	)
	return accepted
}

func ProcessPart(wfMap map[string]*Workflow, part *Part) bool {
	cur := "in"
	for cur != Accept && cur != Reject {
		cur = wfMap[cur].Process(part)
	}
	return cur == Accept
}

type Part struct {
	Values map[byte]int
	Total  int
}

func (p Part) String() string {
	parts := make([]string, 4)
	for i, b := range []byte("xmas") {
		parts[i] = fmt.Sprintf("%c=%d", b, p.Values[b])
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ","))
}

func ParsePart(part string) (*Part, error) {
	if !strings.HasPrefix(part, "{") || !strings.HasSuffix(part, "}") {
		return nil, fmt.Errorf("unknown part format %q", part)
	}

	rv := &Part{Values: make(map[byte]int, 4)}
	entries := strings.Split(strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}"), ",")
	for _, entry := range entries {
		ep := strings.Split(entry, "=")
		if len(ep) != 2 || len(ep[0]) != 1 {
			return nil, fmt.Errorf("failed to parse entry %q from part %q", entry, part)
		}

		val, err := strconv.Atoi(ep[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse part value %q from entry %q in part %q: %w",
				ep[1], entry, part, err)
		}

		rv.Values[ep[0][0]] = val
		rv.Total += val
	}

	return rv, nil
}

type Workflow struct {
	Name  string
	Rules []*Rule
}

func (w Workflow) String() string {
	return fmt.Sprintf("%s{%s}", w.Name, strings.Join(MapSlice(w.Rules, (*Rule).String), ","))
}

var WorkflowRx = regexp.MustCompile(`^([[:alpha:]]+){([^}]+)}$`)

func ParseWorkflow(workflow string) (*Workflow, error) {
	parts := WorkflowRx.FindStringSubmatch(workflow)
	if len(parts) != 3 {
		return nil, fmt.Errorf("failed to parse workflow %q", workflow)
	}

	rv := &Workflow{Name: parts[1]}
	rules := strings.Split(parts[2], ",")
	rv.Rules = make([]*Rule, len(rules))
	var err error
	for i, rule := range rules {
		rv.Rules[i], err = ParseRule(rule)
		if err != nil {
			return nil, err
		}
	}

	return rv, nil
}

func (w Workflow) Process(part *Part) string {
	for _, rule := range w.Rules {
		if rule.AppliesTo(part) {
			return rule.Result
		}
	}

	return ""
}

type Rule struct {
	Field  byte
	Comp   byte
	Val    int
	Result string
}

func (r Rule) String() string {
	if r.Comp == '=' {
		return r.Result
	}
	return fmt.Sprintf("%c%c%d:%s", r.Field, r.Comp, r.Val, r.Result)
}

var RuleRx = regexp.MustCompile(`^([[:alpha:]]+)(<|>)([[:digit:]]+):([[:alpha:]]+)$`)

func ParseRule(rule string) (*Rule, error) {
	parts := RuleRx.FindStringSubmatch(rule)
	if len(parts) != 5 {
		return &Rule{Comp: '=', Result: rule}, nil
	}

	val, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse value %q from rule %q: %w", parts[3], rule, err)
	}

	return &Rule{
		Field:  parts[1][0],
		Comp:   parts[2][0],
		Val:    val,
		Result: parts[4],
	}, nil
}

func (r Rule) AppliesTo(part *Part) bool {
	pval := part.Values[r.Field]
	switch r.Comp {
	case '<':
		return pval < r.Val
	case '>':
		return pval > r.Val
	case '=':
		return true
	default:
		panic(fmt.Errorf("unknown rule comp: %c in %s", r.Comp, r))
	}
}

type Input struct {
	Workflows []*Workflow
	Parts     []*Part
}

func (i Input) String() string {
	return fmt.Sprintf("Workflows (%d):\n%s\n\nParts (%d):\n%s",
		len(i.Workflows), StringNumberJoin(i.Workflows, 1, "\n"),
		len(i.Parts), StringNumberJoin(i.Parts, 1, "\n"),
	)
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{}
	var inParts bool
	for _, line := range lines {
		if len(line) == 0 {
			inParts = true
			continue
		}
		if inParts {
			part, err := ParsePart(line)
			if err != nil {
				return nil, err
			}
			rv.Parts = append(rv.Parts, part)
			continue
		}
		wf, err := ParseWorkflow(line)
		if err != nil {
			return nil, err
		}
		rv.Workflows = append(rv.Workflows, wf)
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
	return strings.Join(AddLineNumbers(MapSlice(slice, E.String), startAt), sep)
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
