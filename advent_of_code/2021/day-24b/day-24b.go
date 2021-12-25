package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const DEFAULT_COUNT = 0

const WXYZ = "wxyz"
const NULLB byte = 0x00

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	Debugf("Parsed Input:\n%s", input)
	nums, err := ParseCustomLines(params.Custom)
	if err != nil {
		return "", err
	}
	answer := 42
	switch params.Count {
	case 4:
		alu := NewALU(input.Prog)
		states, err := alu.Run(nums)
		if err != nil {
			return "", err
		}
		Stdout("Output:\n%s", alu)
		answer = states[len(states)-1]['z']
	case 1:
		alu := NewALU(input.Prog)
		states, err := alu.Run(nums)
		if err != nil {
			return "", err
		}
		Stdout("Output:\n%s", alu)
		CheckDafunc(nums, states)
	case 2:
		alu := NewALU(input.Prog)
		states, err := alu.Run(nums)
		if err != nil {
			return "", err
		}
		Stdout("Output:\n%s", alu)
		tryIt := TryShit()
		if len(tryIt) > 0 {
			alu = NewALU(input.Prog)
			states, err = alu.Run(tryIt)
			if err != nil {
				return "", err
			}
		}
		answer = states[len(states)-1]['z']
	case 3:
		AnalyzeProg(input.Prog)
	case 0:
		chunks := SplitProg(input.Prog)
		wAs, err := BruteForceW4(chunks)
		if err != nil {
			return "", err
		}
		wABs, err := BruteForceW7(chunks, wAs)
		if err != nil {
			return "", err
		}
		wABCs, err := BruteForceW9(chunks, wABs)
		if err != nil {
			return "", err
		}
		wABCDs, err := BruteForceW12(chunks, wABCs)
		if err != nil {
			return "", err
		}
		ws, err := BruteForceW14(chunks, wABCDs)
		if err != nil {
			return "", err
		}
		return StrJoin("", ws[len(ws)-1]), nil
	default:
		return "", fmt.Errorf("unknown --count option: %d", params.Count)
	}
	return fmt.Sprintf("%d", answer), nil
}

func BruteForceW14(chunks ProgramList, w12s [][]int) ([][]int, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	prog12 := Program{}
	prog12 = append(prog12, chunks[0]...)
	prog12 = append(prog12, chunks[1]...)
	prog12 = append(prog12, chunks[2]...)
	prog12 = append(prog12, chunks[3]...)
	prog12 = append(prog12, chunks[4]...)
	prog12 = append(prog12, chunks[5]...)
	prog12 = append(prog12, chunks[6]...)
	prog12 = append(prog12, chunks[7]...)
	prog12 = append(prog12, chunks[8]...)
	prog12 = append(prog12, chunks[9]...)
	prog12 = append(prog12, chunks[10]...)
	prog12 = append(prog12, chunks[11]...)
	var err error
	rv := [][]int{}
	for _, w12 := range w12s {
		w12State := NewProgState(0, 0, 0, 0)
		err = prog12.Run(w12State, w12)
		if err != nil {
			return rv, fmt.Errorf("chunk 12, w: %s: %w", StrJoin("", w12), err)
		}
		for w13 := 9; w13 >= 1; w13-- {
			w13State := w12State.CopyOf()
			err = chunks[12].Run(w13State, []int{w13})
			if err != nil {
				return rv, fmt.Errorf("chunk 13, w: %s: %w", StrJoin("", w12, w13), err)
			}
			if w13State['x'] == 0 {
				Debugf("%s : %s to %s", StrJoin("", w12, w13), w12State, w13State)
				for w14 := 9; w14 >= 1; w14-- {
					w14State := w13State.CopyOf()
					err = chunks[13].Run(w14State, []int{w14})
					if err != nil {
						return rv, fmt.Errorf("chunk 14, w: %s: %w", StrJoin("", w12, w13, w14), err)
					}
					if w14State['x'] == 0 {
						Debugf("%s: %s to %s", StrJoin("", w12, w13, w14), w13State, w14State)
						rv = append(rv, CopyAppend(w12, w13, w14))
					}
				}
			}
		}
	}
	Debugf("Found %d digit combos.", len(rv))
	return rv, nil
}

func BruteForceW12(chunks ProgramList, w9s [][]int) ([][]int, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	prog9 := Program{}
	prog9 = append(prog9, chunks[0]...)
	prog9 = append(prog9, chunks[1]...)
	prog9 = append(prog9, chunks[2]...)
	prog9 = append(prog9, chunks[3]...)
	prog9 = append(prog9, chunks[4]...)
	prog9 = append(prog9, chunks[5]...)
	prog9 = append(prog9, chunks[6]...)
	prog9 = append(prog9, chunks[7]...)
	prog9 = append(prog9, chunks[8]...)
	var err error
	rv := [][]int{}
	for _, w9 := range w9s {
		w9State := NewProgState(0, 0, 0, 0)
		err = prog9.Run(w9State, w9)
		if err != nil {
			return rv, fmt.Errorf("chunk 9, w: %s: %w", StrJoin("", w9), err)
		}
		for w10 := 9; w10 >= 1; w10-- {
			w10State := w9State.CopyOf()
			err = chunks[9].Run(w10State, []int{w10})
			if err != nil {
				return rv, fmt.Errorf("chunk 10, w: %s: %w", StrJoin("", w9, w10), err)
			}
			for w11 := 9; w11 >= 1; w11-- {
				w11State := w10State.CopyOf()
				err = chunks[10].Run(w11State, []int{w11})
				if err != nil {
					return rv, fmt.Errorf("chunk 11, w: %s: %w", StrJoin("", w9, w10, w11), err)
				}
				for w12 := 9; w12 >= 1; w12-- {
					w12State := w11State.CopyOf()
					err = chunks[11].Run(w12State, []int{w12})
					if err != nil {
						return rv, fmt.Errorf("chunk 12, w: %s: %w", StrJoin("", w9, w10, w11, w12), err)
					}
					if w12State['x'] == 0 {
						Debugf("%s: %s to %s", StrJoin("", w9, w10, w11, w12), w11State, w12State)
						rv = append(rv, CopyAppend(w9, w10, w11, w12))
					}
				}
			}
		}
	}
	Debugf("Found %d first 12 digit combos.", len(rv))
	return rv, nil
}

func BruteForceW9(chunks ProgramList, w7s [][]int) ([][]int, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	prog7 := Program{}
	prog7 = append(prog7, chunks[0]...)
	prog7 = append(prog7, chunks[1]...)
	prog7 = append(prog7, chunks[2]...)
	prog7 = append(prog7, chunks[3]...)
	prog7 = append(prog7, chunks[4]...)
	prog7 = append(prog7, chunks[5]...)
	prog7 = append(prog7, chunks[6]...)
	var err error
	rv := [][]int{}
	for _, w7 := range w7s {
		w7State := NewProgState(0, 0, 0, 0)
		err = prog7.Run(w7State, w7)
		if err != nil {
			return rv, fmt.Errorf("chunk 7, w: %s: %w", StrJoin("", w7), err)
		}
		for w8 := 9; w8 >= 1; w8-- {
			w8State := w7State.CopyOf()
			err = chunks[7].Run(w8State, []int{w8})
			if err != nil {
				return rv, fmt.Errorf("chunk 8, w: %s: %w", StrJoin("", w7, w8), err)
			}
			for w9 := 9; w9 >= 1; w9-- {
				w9State := w8State.CopyOf()
				err = chunks[8].Run(w9State, []int{w9})
				if err != nil {
					return rv, fmt.Errorf("chunk 9, w: %s: %w", StrJoin("", w7, w8, w9), err)
				}
				if w9State['x'] == 0 {
					Debugf("%s: %s to %s", StrJoin("", w7, w8, w9), w8State, w9State)
					rv = append(rv, CopyAppend(w7, w8, w9))
				}
			}
		}
	}
	Debugf("Found %d first 9 digit combos.", len(rv))
	return rv, nil
}

func BruteForceW7(chunks ProgramList, w4s [][]int) ([][]int, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	prog4 := Program{}
	prog4 = append(prog4, chunks[0]...)
	prog4 = append(prog4, chunks[1]...)
	prog4 = append(prog4, chunks[2]...)
	prog4 = append(prog4, chunks[3]...)
	var err error
	rv := [][]int{}
	for _, w4 := range w4s {
		w4State := NewProgState(0, 0, 0, 0)
		err = prog4.Run(w4State, w4)
		if err != nil {
			return rv, fmt.Errorf("chunk 4, w: %s: %w", StrJoin("", w4), err)
		}
		for w5 := 9; w5 >= 1; w5-- {
			w5State := w4State.CopyOf()
			err = chunks[4].Run(w5State, []int{w5})
			if err != nil {
				return rv, fmt.Errorf("chunk 5, w: %s: %w", StrJoin("", w4, w5), err)
			}
			for w6 := 9; w6 >= 1; w6-- {
				w6State := w5State.CopyOf()
				err = chunks[5].Run(w6State, []int{w6})
				if err != nil {
					return rv, fmt.Errorf("chunk 6, w: %s: %w", StrJoin("", w4, w5, w6), err)
				}
				if w6State['x'] == 0 {
					Debugf("%s : %s to %s", StrJoin("", w4, w5, w6), w5State, w6State)
					for w7 := 9; w7 >= 1; w7-- {
						w7State := w6State.CopyOf()
						err = chunks[6].Run(w7State, []int{w7})
						if err != nil {
							return rv, fmt.Errorf("chunk 7, w: %s: %w", StrJoin("", w4, w5, w6, w7), err)
						}
						if w7State['x'] == 0 {
							Debugf("%s: %s to %s", StrJoin("", w4, w5, w6, w7), w6State, w7State)
							rv = append(rv, CopyAppend(w4, w5, w6, w7))
						}
					}
				}
			}
		}
	}
	Debugf("Found %d first 7 digit combos.", len(rv))
	return rv, nil
}

func BruteForceW4(chunks ProgramList) ([][]int, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	var err error
	rv := [][]int{}
	for w1 := 9; w1 >= 1; w1-- {
		w1State := NewProgState(0, 0, 0, 0)
		err = chunks[0].Run(w1State, []int{w1})
		if err != nil {
			return rv, fmt.Errorf("chunk 1, w: %d: %w", w1, err)
		}
		for w2 := 9; w2 >= 1; w2-- {
			w2State := w1State.CopyOf()
			err = chunks[1].Run(w2State, []int{w2})
			if err != nil {
				return rv, fmt.Errorf("chunk 2, w: %d %d: %w", w1, w2, err)
			}
			for w3 := 9; w3 >= 1; w3-- {
				w3State := w2State.CopyOf()
				err = chunks[2].Run(w3State, []int{w3})
				if err != nil {
					return rv, fmt.Errorf("chunk 3, w: %d %d %d: %w", w1, w2, w3, err)
				}
				for w4 := 9; w4 >= 1; w4-- {
					w4State := w3State.CopyOf()
					err = chunks[3].Run(w4State, []int{w4})
					if err != nil {
						return rv, fmt.Errorf("chunk 4, w: %d %d %d %d: %w", w1, w2, w3, w4, err)
					}
					if w4State['x'] == 0 {
						Debugf("%d%d%d%d: %s to %s", w1, w2, w3, w4, w3State, w4State)
						rv = append(rv, []int{w1, w2, w3, w4})
					}
				}
			}
		}
	}
	w4s := map[int]bool{}
	for _, ws := range rv {
		// exp := (((ws[0]+6)*26+ws[1]+14)*26+ws[2]+13)%26 - 14 // This is correct.
		exp := (ws[2]+13)%26 - 14 // This is correct too.
		if ws[3] != exp {
			Stdout("(%d+13)%%26 - 14 = %d =?= %d <<<<<< WRONG!", ws[2], exp, ws[3])
		}
		w4s[ws[3]] = true
	}
	notFound := []int{}
	for i := 1; i <= 9; i++ {
		if !w4s[i] {
			notFound = append(notFound, i)
		}
	}
	if len(notFound) == 0 {
		Stdout("w4 can be any digit.")
	} else {
		Stdout("w4 cannot be %v.", notFound)
	}
	Debugf("Found %d first 4 digit combos.", len(rv))
	return rv, nil
}

func CopyAppend(vals []int, more ...int) []int {
	if len(more) == 0 {
		rv := make([]int, len(vals))
		copy(rv, vals)
		return rv
	}
	if len(vals) == 0 {
		rv := make([]int, len(more))
		copy(rv, more)
		return rv
	}
	rv := make([]int, len(vals)+len(more))
	copy(rv, vals)
	copy(rv[len(vals):], more)
	return rv
}

func StrJoin(d string, vals []int, more ...int) string {
	parts := make([]string, len(vals)+len(more))
	for i, val := range vals {
		parts[i] = fmt.Sprintf("%d", val)
	}
	for i, val := range more {
		parts[i+len(vals)] = fmt.Sprintf("%d", val)
	}
	return strings.Join(parts, d)
}

func AnalyzeProg(prog Program) {
	chunks := SplitProg(prog)
	Stdout("There are %d chunks.:", len(chunks))
	for i, chunk := range chunks {
		vals := []string{}
		for l, inst := range chunk {
			if l == 4 || l == 5 || l == 15 {
				vals = append(vals, fmt.Sprintf("% 3d", inst.Val2))
			}
		}
		Stdout("%2d: Instruction Count: %d, Vals: %s", i, len(chunk), strings.Join(vals, ", "))
	}
	diffFound := false
	isNotableDiff := func(c, l int) bool {
		if chunks[c][l].T != chunks[0][l].T {
			return true
		}
		if chunks[c][l].Var1 != chunks[0][l].Var1 {
			return true
		}
		if l != 4 && l != 5 && l != 15 && (chunks[c][l].Var2 != chunks[0][l].Var2 || chunks[c][l].Val2 != chunks[0][l].Val2) {
			return true
		}
		return false
	}
	for l := 0; l < 18; l++ {
		for c := 1; c < 14; c++ {
			if isNotableDiff(c, l) {
				Stdout("Chunk %d line %d = %s  vs  %s in chunk 0", c+1, l+1, chunks[c][l])
				diffFound = true
			}
		}
	}
	if !diffFound {
		Stdout("No differences found in instruction type and var1 across chunks.")
	}
}

type ProgramList []Program

func (l ProgramList) String() string {
	lines := []string{}
	for i, prog := range l {
		for _, line := range strings.Split(prog.String(), "\n") {
			lines = append(lines, fmt.Sprintf("%2d: %s\n", i+1, line))
		}
	}
	return strings.Join(lines, "\n")
}

func SplitProg(prog Program) ProgramList {
	rv := ProgramList{}
	cur := -1
	for _, inst := range prog {
		if inst.T == INP {
			cur++
			rv = append(rv, Program{inst})
		} else {
			rv[cur] = append(rv[cur], inst)
		}
	}
	return rv
}

func TryShit() []int {
	As := []int{8, 9, 10, 11, 12, 13, 14, 15, 16}
	Bs := []int{4, 5, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	Cs := []int{6, 7, 8, 9, 10, 11, 12, 13, 14, 32, 33, 34, 35, 36, 37, 38, 39, 40}
	Ds := []int{7, 8, 9, 10, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62}
	isWOK := func(w int) bool {
		return w >= 1 && w <= 9
	}
	Debugf("Making W1 through W4 options.")
	wAs := [][]int{}
	for w1 := 9; w1 >= 1; w1-- {
		for w2 := 9; w2 >= 1; w2-- {
			for w3 := 9; w3 >= 1; w3-- {
				if w1+w2+w3 >= As[0] && w1+w2+w3 <= As[len(As)-1] {
					w4 := (w1+w2+w3+7)%26 - 14
					if isWOK(w4) {
						wAs = append(wAs, []int{w1, w2, w3, w4})
					}
				}
			}
		}
	}
	Debugf("There are %d W1 through W4 options.", len(wAs))
	// Reverse the Bs to get the biggest first.
	for i, j := 0, len(Bs)-1; i < j; i, j = i+1, j-1 {
		Bs[i], Bs[j] = Bs[j], Bs[i]
	}
	Debugf("Making W1 through W7 options.")
	wBs := [][]int{}
	for _, w := range wAs {
		a := w[0] + w[1] + w[2]
		for _, b := range Bs {
			w5 := b - a
			if isWOK(w5) {
				w6 := (b + 10) % 26
				w7 := (b+10)%26 - 6
				if isWOK(w6) && isWOK(w7) {
					wBs = append(wBs, []int{w[0], w[1], w[2], w[3], w5, w6, w7})
				}
			}
		}
	}
	Debugf("There are %d W1 through W7 options.", len(wBs))
	// Reverse the Cs to get the biggest first.
	for i, j := 0, len(Cs)-1; i < j; i, j = i+1, j-1 {
		Cs[i], Cs[j] = Cs[j], Cs[i]
	}
	Debugf("Making W1 through W9 options.")
	wCs := [][]int{}
	for _, w := range wBs {
		b := w[0] + w[1] + w[2] + w[4]
		for _, c := range Cs {
			w8 := c - b
			if isWOK(w8) {
				w9 := (c+24)%26 - 3
				if isWOK(w9) {
					wCs = append(wCs, []int{w[0], w[1], w[2], w[3], w[4], w[5], w[6], w8, w9})
				}
			}
		}
	}
	Debugf("There are %d W1 through W9 options.", len(wCs))
	// Reverse the Ds to get the biggest first.
	for i, j := 0, len(Ds)-1; i < j; i, j = i+1, j-1 {
		Ds[i], Ds[j] = Ds[j], Ds[i]
	}
	Debugf("Making W1 through W14 options.")
	wDs := [][]int{}
	for _, w := range wCs {
		c := w[0] + w[1] + w[2] + w[4] + w[7]
		for _, d := range Ds {
			w1011 := d - c
			for w10 := 9; w10 >= 1; w10-- {
				w11 := w1011 - w10
				if isWOK(w10) && isWOK(w11) {
					w12 := (d+8)%26 - 2
					w13 := (d+8)%26 - 9
					if isWOK(w12) && isWOK(w13) {
						wDs := append(wDs, make([]int, 14))
						copy(wDs[len(wDs)-1], w)
						copy(wDs[len(wDs)-1][len(w):], []int{w10, w11, w12, w13, w12})
					}
				}
			}
		}
	}
	Debugf("There are %d W1 through W14 options.", len(wDs))
	if debug {
		lines := make([]string, len(wDs))
		for i, try := range wDs {
			lines[i] = strings.Trim(strings.ReplaceAll(fmt.Sprintf("%v", try), " ", ""), "[]")
		}
		Stderr("Here they are:\n%s", strings.Join(AddLineNumbers(lines, 1), "\n"))
	}
	if len(wDs) == 0 {
		return []int{}
	}
	return wDs[0]
	// Clearly my math is wrong somewhere.
	// (   0.011369542)   [TryShit] Making W1 through W4 options.
	// (   0.011502250)   [TryShit] There are 420 W1 through W4 options.
	// (   0.011516459)   [TryShit] Making W1 through W7 options.
	// (   0.011653709)   [TryShit] There are 362 W1 through W7 options.
	// (   0.011671500)   [TryShit] Making W1 through W9 options.
	// (   0.011850042)   [TryShit] There are 603 W1 through W9 options.
	// (   0.011864000)   [TryShit] Making W1 through W14 options.
	// (   0.012958459)   [TryShit] There are 0 W1 through W14 options.
	// (   0.012987292)   [TryShit] Here they are:
}

func CheckDafunc(nums []int, exp []ProgState) {
	varss := [][]int{
		{1, 11, 6},
		{1, 11, 14},
		{1, 15, 13},
		{26, -14, 1},
		{1, 10, 6},
		{26, 0, 13},
		{26, -6, 6},
		{1, 13, 3},
		{26, -3, 8},
		{1, 13, 14},
		{1, 15, 4},
		{26, -2, 7},
		{26, -9, 15},
		{26, -2, 1},
	}
	z := 0
	for i, vars := range varss {
		newZ := DaFunc(nums[i], z, vars[0], vars[1], vars[2])
		bad := ""
		if newZ != exp[i]['z'] {
			bad = " <==== WRONG"
		}
		Stdout("%2d: w: %d, z: %d, vars: %v => %d vs %d%s", i+1, nums[i], z, vars, newZ, exp[i]['z'], bad)
		z = newZ
	}
}

func DaFunc(w, z, v1, v2, v3 int) int {
	if z%26+v2 != w {
		return z/v1*26 + w + v3
	}
	return z / v1
}

func ParseCustomLines(lines []string) ([]int, error) {
	if len(lines) == 0 {
		return []int{27, 3}, nil
	}
	rv := []int{}
	for _, line := range lines {
		switch {
		case len(line) == 14:
			for _, b := range line {
				val, err := strconv.Atoi(string(b))
				if err != nil {
					return nil, err
				}
				rv = append(rv, val)
			}
		case len(line) > 0:
			val, err := strconv.Atoi(line)
			if err != nil {
				return nil, err
			}
			rv = append(rv, val)
		}
	}
	return rv, nil
}

type ProgState map[byte]int

func (p ProgState) String() string {
	lines := make([]string, len(WXYZ))
	for i, b := range []byte(WXYZ) {
		lines[i] = fmt.Sprintf("%c = %d", b, p[b])
	}
	return strings.Join(lines, ", ")
}

func (p ProgState) CopyOf() ProgState {
	return map[byte]int{
		'w': p['w'],
		'x': p['x'],
		'y': p['y'],
		'z': p['z'],
	}
}

func NewProgState(w, x, y, z int) ProgState {
	return map[byte]int{
		'w': w,
		'x': x,
		'y': y,
		'z': z,
	}
}

type ALU struct {
	Ptr  int
	Prog Program
	Reg  ProgState
}

func (a ALU) String() string {
	return fmt.Sprintf("Ptr: %d of %d. State: %s", a.Ptr, len(a.Prog), a.Reg)
}

func NewALU(prog Program) *ALU {
	return &ALU{
		Prog: prog,
		Reg:  NewProgState(0, 0, 0, 0),
	}
}

func NewALUWithState(prog Program, state ProgState) *ALU {
	return &ALU{
		Prog: prog,
		Reg:  state,
	}
}

func (a *ALU) Run(input []int) ([]ProgState, error) {
	Stdout("Input Values: %v", input)
	var err error
	rv := []ProgState{}
	for a.Ptr < len(a.Prog) {
		if a.Prog[a.Ptr].T == INP {
			rv = append(rv, a.Reg.CopyOf())
		}
		input, err = a.Run1(input)
		if err != nil {
			return rv, err
		}
	}
	rv = append(rv, a.Reg.CopyOf())
	return rv, nil
}

func (a *ALU) Run1(input []int) ([]int, error) {
	inst := a.Prog[a.Ptr]
	var err error
	input, err = inst.Execute(a.Reg, input)
	a.Ptr++
	if err != nil {
		return input, fmt.Errorf("error on line %d: %w", a.Ptr, err)
	}
	return input, nil
}

type InstType int

const INP InstType = 1
const ADD InstType = 2
const MUL InstType = 3
const DIV InstType = 4
const MOD InstType = 5
const EQL InstType = 6

func (i InstType) String() string {
	switch i {
	case INP:
		return "inp"
	case ADD:
		return "add"
	case MUL:
		return "mul"
	case DIV:
		return "div"
	case MOD:
		return "mod"
	case EQL:
		return "eql"
	}
	return fmt.Sprintf("InstType(%d)", i)
}

type Instruction struct {
	T    InstType
	Var1 byte
	Var2 byte
	Val2 int
}

func (i Instruction) String() string {
	if i.T == INP {
		return fmt.Sprintf("%s %c", i.T, i.Var1)
	}
	if i.Var2 == NULLB {
		return fmt.Sprintf("%s %c %d", i.T, i.Var1, i.Val2)
	}
	return fmt.Sprintf("%s %c %c", i.T, i.Var1, i.Var2)
}

func ParseInstruction(str string) (*Instruction, error) {
	if len(str) < 5 {
		return nil, fmt.Errorf("invalid instruction: too short: %q", str)
	}
	rv := Instruction{}
	switch str[0:3] {
	case "inp":
		rv.T = INP
	case "add":
		rv.T = ADD
	case "mul":
		rv.T = MUL
	case "div":
		rv.T = DIV
	case "mod":
		rv.T = MOD
	case "eql":
		rv.T = EQL
	}
	rv.Var1 = str[4]
	if rv.T != INP {
		if len(str) < 7 {
			return nil, fmt.Errorf("invalid instruction: expecting second value: %q", str)
		}
		if bytes.ContainsAny([]byte{str[6]}, WXYZ) {
			rv.Var2 = str[6]
		} else {
			var err error
			rv.Val2, err = strconv.Atoi(str[6:])
			if err != nil {
				return nil, fmt.Errorf("invalid instruction: could not parse [%s] to a number from %q: %w", str[6:], str, err)
			}
		}
	}
	return &rv, nil
}

func (i Instruction) Execute(reg ProgState, input []int) ([]int, error) {
	if i.T == INP {
		if len(input) == 0 {
			return input, fmt.Errorf("need input: %s", i)
		}
		reg[i.Var1] = input[0]
		return input[1:], nil
	}
	val2 := i.Val2
	if i.Var2 != NULLB {
		val2 = reg[i.Var2]
	}
	switch i.T {
	case ADD:
		reg[i.Var1] += val2
	case MUL:
		reg[i.Var1] *= val2
	case DIV:
		if val2 == 0 {
			return input, fmt.Errorf("invalid div values: %s, %d / %d", i, reg[i.Var1], val2)
		}
		reg[i.Var1] = reg[i.Var1] / val2
	case MOD:
		if reg[i.Var1] < 0 || val2 <= 0 {
			return input, fmt.Errorf("invalid mod values: %s, %d / %d", i, reg[i.Var1], val2)
		}
		reg[i.Var1] = reg[i.Var1] % val2
	case EQL:
		if reg[i.Var1] == val2 {
			reg[i.Var1] = 1
		} else {
			reg[i.Var1] = 0
		}
	}
	return input, nil
}

type Program []*Instruction

func (p Program) String() string {
	lines := make([]string, len(p))
	for i, inst := range p {
		lines[i] = inst.String()
	}
	return strings.Join(AddLineNumbers(lines, 1), "\n")
}

func (p Program) Run(reg ProgState, input []int) error {
	var err error
	for _, inst := range p {
		input, err = inst.Execute(reg, input)
		if err != nil {
			return err
		}
	}
	return nil
}

type Input struct {
	Prog Program
}

func (i Input) String() string {
	return fmt.Sprintf("Program:\n%s", i.Prog)
}

func ParseInput(lines []string) (*Input, error) {
	rv := Input{}
	for i, line := range lines {
		if len(line) > 0 {
			inst, err := ParseInstruction(line)
			if err != nil {
				return &rv, fmt.Errorf("could not parse line %d: %w", i+1, err)
			}
			rv.Prog = append(rv.Prog, inst)
		}
	}
	return &rv, nil
}

// -------------------------------------------------------------------------------------
// -------------------------------  Some generic stuff  --------------------------------
// -------------------------------------------------------------------------------------

const MIN_INT8 = int8(-128)
const MAX_INT8 = int8(127)
const MIN_INT16 = int16(-32768)
const MAX_INT16 = int16(32767)
const MIN_INT32 = int32(-2147483648)
const MAX_INT32 = int32(2147483647)
const MIN_INT64 = int64(-9223372036854775808)
const MAX_INT64 = int64(9223372036854775807)
const MIN_INT = -9223372036854775808
const MAX_INT = 9223372036854775807

const MAX_UINT8 = uint8(255)
const MAX_UINT16 = uint16(65535)
const MAX_UINT32 = uint32(4294967295)
const MAX_UINT64 = uint64(18446744073709551615)
const MAX_UINT = uint(18446744073709551615)

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
// E.g. DigitFormatForMax(10) returns "%2d"
// DigitFormatForMax(382920) returns "%6d"
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

// -------------------------------------------------------------------------------------
// --------------------------  CLI params and input parsing  ---------------------------
// -------------------------------------------------------------------------------------

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

// String creates a multi-line string representing this Params
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

// DEFAULT_INPUT_FILE is the default input filename
const DEFAULT_INPUT_FILE = "actual.input"

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
			// Using fmt.Println here instead of my stdout function because the extra formatting is annoying with help text.
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
					Stderr("Debugging enabled by CLI arguments.")
				case oldDebug && !debug:
					Stderr("Debugging disabled by CLI arguments.")
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
		lines := []string{fmt.Sprintf("Found %d errors:", len(c.Errors))}
		for i, err := range c.Errors {
			lines = append(lines, fmt.Sprintf("  %d: %s", i, err.Error()))
		}
		return errors.New(strings.Join(lines, "\n"))
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
		if (len(arg) == 2 && arg[0] == '-') || strings.HasPrefix(arg, "--") {
			return rv, len(rv), nil
		}
		rv = append(rv, arg)
	}
	if len(rv) >= 0 {
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
	DebugfAlways("Reading file: %s", filename)
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		Stderr("error reading file: %v", err)
		return []string{}, err
	}
	return strings.Split(string(dat), "\n"), nil
}

// -------------------------------------------------------------------------------------
// --------------------------  Environment Variable Handling  --------------------------
// -------------------------------------------------------------------------------------

// HandleEnvVars looks at specific environment variables and sets global variables appropriately.
func HandleEnvVars() error {
	var err error
	debug, err = GetEnvVarBool("DEBUG")
	if debug {
		Stderr("Debugging enabled via environment variable.")
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

// -------------------------------------------------------------------------------------
// ------------------------  Function start/stop timing stuff  -------------------------
// -------------------------------------------------------------------------------------

// If all you want is starting/ending messages when debug is on, use:
//    defer FuncEnding(FuncStarting())
// If, when debug is on, you want starting/ending messages,
// but when debug is off, you still the function duration, then use:
//    defer FuncEndingAlways(FuncStarting())

// FuncStarting outputs that a function is starting (if debug is true).
// It returns the params needed by FuncEnding or FuncEndingAlways.
//
// Arguments provided will be converted to stings using %v and included as part of the function name.
// Minimal values needed to differentiate start/stop output lines should be provided.
// Long strings and complex structs should be avoided.
//
// Example 1: In a function named "foo", you have this:
//     FuncStarting()
//   The printed message will note that "foo" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 2: In a function named "bar", you have this:
//     FuncStarting(3 * time.Second)
//   The printed message will note that "bar: 3s" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Example 3:
//     func sum(ints ...int) {
//         FuncStarting(ints...)
//     }
//     sum(1, 2, 3, 4, 20, 21, 22)
//   The printed message will note that "sum: 1, 2, 3, 4, 20, 21, 22" is starting.
//   That same string will also be returned as the 2nd return paremeter.
//
// Standard Usage: defer FuncEnding(FuncStarting())
//             Or: defer FuncEndingAlways(FuncStarting())
func FuncStarting(a ...interface{}) (time.Time, string) {
	funcDepth++
	name := GetFuncName(1, a...)
	DebugfAs(name, "Starting.")
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
	DebugfAlwaysAs(name, "Starting.")
	return time.Now(), name
}

const done_fmt = "Done. Duration: [%s]."

// FuncEnding decrements the function depth and, if debug is on, outputs to stderr that how long a function took.
// Args will usually come from FuncStarting().
//
// This differs from FuncEndingAlways in that this only outputs something if debugging is turned on.
//
// Standard Usage: defer FuncEnding(FuncStarting())
func FuncEnding(start time.Time, name string) {
	DebugfAs(name, done_fmt, time.Since(start))
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
	DebugfAlwaysAs(name, done_fmt, time.Since(start))
	if funcDepth > -1 {
		funcDepth--
	}
}

// DurClock converts a duration to a string in minimal clock notation with nanosecond precision.
//
// - If one or more hours, format is "H:MM:SS.NNNNNNNNNs", e.g. "12:01:02.000000000"
// - If less than one hour, format is "M:SS.NNNNNNNNNs",   e.g. "34:00.000000789"
// - If less than one minute, format is "S.NNNNNNNNNs",    e.g. "56.000456000"
// - If less than one second, format is "0.NNNNNNNNNs",    e.g. "0.123000000"
func DurClock(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes())
	s := int(d.Seconds())
	n := int(d.Nanoseconds()) - 1000000000*s
	s = s - 60*m
	m = m - 60*h
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
// depth 0 = the function calling GetFuncName.
// depth 1 = the function calling the function calling GetFuncName.
// etc.
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

// -------------------------------------------------------------------------------------
// ---------------------------------  Output wrappers  ---------------------------------
// -------------------------------------------------------------------------------------

// GetOutputPrefix gets the prefix to add to all output.
func GetOutputPrefix(funcName string) string {
	tabs := ""
	if debug && funcDepth > 0 {
		tabs = strings.Repeat("  ", funcDepth)
	}
	return fmt.Sprintf("(%14s) %s[%s] ", DurClock(time.Since(startTime)), tabs, funcName)
}

// Stdout outputs to stdout with a prefixed run duration and automatic function name.
func Stdout(format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// Stderr outputs to stderr with a prefixed run duration and automatic function name.
func Stderr(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(GetFuncName(1))+format+"\n", a...)
}

// StdoutAs outputs to stdout with a prefixed run duration and provided function name.
func StdoutAs(funcName, format string, a ...interface{}) {
	fmt.Printf(GetOutputPrefix(funcName)+format+"\n", a...)
}

// StderrAs outputs to stderr with a prefixed run duration and provided functio name.
func StderrAs(funcName, format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, GetOutputPrefix(funcName)+format+"\n", a...)
}

// Debugf outputs to stderr if the debug flag is set.
func Debugf(format string, a ...interface{}) {
	if debug {
		StderrAs(GetFuncName(1), format, a...)
	}
}

// DebugfAs outputs to stderr if the debug flag is set.
func DebugfAs(funcName, format string, a ...interface{}) {
	if debug {
		StderrAs(funcName, format, a...)
	}
}

// DebugfAlways outputs to stderr if the debug flag is set, otherwise to stdout.
func DebugfAlways(format string, a ...interface{}) {
	if debug {
		StderrAs(GetFuncName(1), format, a...)
	} else {
		StdoutAs(GetFuncName(1), format, a...)
	}
}

// DebugfAlways outputs to stderr if the debug flag is set, otherwise to stdout.
func DebugfAlwaysAs(funcName, format string, a ...interface{}) {
	if debug {
		StderrAs(funcName, format, a...)
	} else {
		StdoutAs(funcName, format, a...)
	}
}

// -------------------------------------------------------------------------------------
// --------------------------  Primary Program Running Parts  --------------------------
// -------------------------------------------------------------------------------------

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
		// Not using Stderr(...) here because I don't want the time and function prefix on this.
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
	Stdout("Answer: %s", answer)
	return nil
}
