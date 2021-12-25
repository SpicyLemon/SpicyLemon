package main

import (
	"bytes"
	"container/heap"
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

var Cost = map[byte]int{
	'A': 1,
	'B': 10,
	'C': 100,
	'D': 1000,
}

var AmphHome = map[byte]int{
	'A': 0,
	'B': 1,
	'C': 2,
	'D': 3,
}

var RoomAmph = map[int]byte{
	0: 'A',
	1: 'B',
	2: 'C',
	3: 'D',
}

var CompletedBurrow = &Burrow{
	Hall: []byte("......."),
	Rooms: [][]byte{
		[]byte("AAAA"),
		[]byte("BBBB"),
		[]byte("CCCC"),
		[]byte("DDDD"),
	},
}

// Solve is the main entry point to finding a solution.
// The string it returns should be (or include) the answer.
func Solve(params *Params) (string, error) {
	defer FuncEndingAlways(FuncStartingAlways())
	input, err := ParseInput(params.Input)
	if err != nil {
		return "", err
	}
	Debugf("Parsed Input:\n%s", input)
	burrow := input.Start
	answer := 0
	answer += BurrowMustMove(burrow, RoomSpot(1, 0), HallSpot(6)) // Could have just moved this to HallSpot(5). Saves 1 step = 100.
	answer += BurrowMustMove(burrow, RoomSpot(1, 1), HallSpot(5)) // and this to just HallSpot(4). Saves 2 steps = 200.
	answer += BurrowMustMove(burrow, RoomSpot(1, 2), HallSpot(3))
	answer += BurrowMustMove(burrow, RoomSpot(1, 3), HallSpot(0))
	answer += BurrowMustMove(burrow, HallSpot(3), RoomSpot(1, 3))
	answer += BurrowMustMove(burrow, RoomSpot(2, 0), RoomSpot(1, 2))
	answer += BurrowMustMove(burrow, RoomSpot(2, 1), RoomSpot(1, 1))
	answer += BurrowMustMove(burrow, RoomSpot(2, 2), HallSpot(1))
	answer += BurrowMustMove(burrow, RoomSpot(2, 3), HallSpot(3))
	answer += BurrowMustMove(burrow, RoomSpot(3, 0), RoomSpot(2, 3))
	answer += BurrowMustMove(burrow, HallSpot(5), RoomSpot(2, 2)) // Then this comes from HallSpot(4). Saves 2 steps = 200.
	answer += BurrowMustMove(burrow, HallSpot(6), RoomSpot(2, 1)) // And this from HallSpot(5). Saves 1 step = 100.
	answer += BurrowMustMove(burrow, RoomSpot(3, 1), HallSpot(6))
	answer += BurrowMustMove(burrow, RoomSpot(3, 2), RoomSpot(2, 0))
	answer += BurrowMustMove(burrow, RoomSpot(3, 3), HallSpot(5))
	answer += BurrowMustMove(burrow, HallSpot(3), RoomSpot(3, 3))
	answer += BurrowMustMove(burrow, HallSpot(5), RoomSpot(1, 0))
	answer += BurrowMustMove(burrow, RoomSpot(0, 0), HallSpot(5))
	answer += BurrowMustMove(burrow, RoomSpot(0, 1), RoomSpot(3, 2))
	answer += BurrowMustMove(burrow, RoomSpot(0, 2), RoomSpot(3, 1))
	answer += BurrowMustMove(burrow, RoomSpot(0, 3), RoomSpot(3, 0))
	answer += BurrowMustMove(burrow, HallSpot(1), RoomSpot(0, 3))
	answer += BurrowMustMove(burrow, HallSpot(0), RoomSpot(0, 2))
	answer += BurrowMustMove(burrow, HallSpot(5), RoomSpot(0, 1))
	answer += BurrowMustMove(burrow, HallSpot(6), RoomSpot(0, 0))
	// This solution costs 49276, which is 600 higher than the min, accounted for in the comments on a few lines above.
	return fmt.Sprintf("%d", answer), nil
}

func GetMoves(node *PQNode) MoveNodeList {
	rv := MoveNodeList{node.Ele}
	cur := node.ComeFrom
	for cur != nil {
		rv = append(rv, cur.Ele)
		cur = cur.ComeFrom
	}
	for i, j := 0, len(rv)-1; i < j; i, j = i+1, j-1 {
		rv[i], rv[j] = rv[j], rv[i]
	}
	return rv
}

type Solver struct {
	Visited   []*PQNode
	Unvisited PriorityQueue
	Solution  *PQNode
}

func (s Solver) String() string {
	return fmt.Sprintf("Unvisited: %d, Visited: %d, Solution: %s", len(s.Unvisited), len(s.Visited), s.Solution)
}

func NewSolver(burrow *Burrow) *Solver {
	firstMoves := GetNextMoves(burrow)
	rv := Solver{
		Unvisited: make(PriorityQueue, len(firstMoves)),
		Visited: []*PQNode{
			&PQNode{
				Ele: &MoveNode{
					State: burrow,
				},
				MinCost: 0,
				Index:   -1,
				Visited: true,
			},
		},
	}
	for i, move := range firstMoves {
		rv.Unvisited[i] = &PQNode{
			Ele:     move,
			MinCost: move.Cost,
			Index:   i,
			Visited: false,
		}
	}
	heap.Init(&rv.Unvisited)
	Debugf("Initial list of possible moves:\n%s", firstMoves)
	return &rv
}

func (s *Solver) CalculateNext() {
	defer FuncEnding(FuncStarting())
	keyNode := heap.Pop(&s.Unvisited).(*PQNode)
	// Debugf("Key Node: %s", keyNode)
	nextMoves := GetNextMoves(keyNode.Ele.State)
	if debug {
		Stderr("Calculating Next with Key Node: %s\n%s\nNext Moves To Consider:\n%s", keyNode, keyNode.Ele.State, nextMoves)
	}
	for _, move := range nextMoves {
		Debugf("Checking if this is complete:\n%s", move.State)
		if CompletedBurrow.Equals(move.State) {
			Debugf("Solution found:\n%s", move)
			s.Solution = &PQNode{
				Ele:      move,
				MinCost:  keyNode.MinCost + move.Cost,
				ComeFrom: keyNode,
			}
			break
		}
		found := false
		for _, node := range s.Visited {
			if move.State.Equals(node.Ele.State) {
				found = true
				break
			}
		}
		if found {
			continue
		}
		for _, node := range s.Unvisited {
			if move.State.Equals(node.Ele.State) {
				found = true
				if keyNode.MinCost+move.Cost < node.MinCost {
					s.Unvisited.Update(node, keyNode.MinCost+move.Cost, keyNode)
				}
				break
			}
		}
		if found {
			continue
		}
		node := PQNode{
			Ele:      move,
			MinCost:  keyNode.MinCost + move.Cost,
			ComeFrom: keyNode,
		}
		heap.Push(&s.Unvisited, &node)
	}
	s.Visited = append(s.Visited, keyNode)
}

func (s Solver) IsDone() bool {
	return s.Solution != nil || len(s.Unvisited) == 0
}

func GetNextMoves(b *Burrow) MoveNodeList {
	movableAmphs := []Spot{}
	emptyHalls := []Spot{}
	emptyRooms := []Spot{}
	// Look at the hall and get all amphs and empty spots.
	for i := range b.Hall {
		if b.Hall[i] == '.' {
			emptyHalls = append(emptyHalls, HallSpot(i))
		} else {
			movableAmphs = append(movableAmphs, HallSpot(i))
		}
	}
	// Look at the rooms and get the top empty and top amph.
	// If there's an amph at the top, there isn't an empty room.
	// If there's an empty room at the bottom, there's no amph.
	for r := range b.Rooms {
		for d := range b.Rooms[r] {
			if b.Rooms[r][d] != '.' {
				if r != AmphHome[b.Rooms[r][d]] {
					movableAmphs = append(movableAmphs, RoomSpot(r, d))
				} else {
					for d2 := d + 1; d2 < len(b.Rooms[r]); d2++ {
						if b.Rooms[r][d] != b.Rooms[r][d2] {
							movableAmphs = append(movableAmphs, RoomSpot(r, d))
						}
					}
				}
				if d > 0 {
					emptyRooms = append(emptyRooms, RoomSpot(r, d-1))
				}
				break
			}
			if d == len(b.Rooms[r])-1 {
				emptyRooms = append(emptyRooms, RoomSpot(r, d))
			}
		}
	}
	// Combine all hall -> room, room -> hall, and room -> room moves.
	rv := MoveNodeList{}
	checkAndAppend := func(move *Move) bool {
		Debugf("Checking on Move: %s", move)
		mn, err := NewMoveNode(b, move)
		if err == nil {
			rv = append(rv, mn)
			return true
		}
		Debugf("Move ignored: %v", err)
		return false
	}
	Debugf("Amph Spots: %s", movableAmphs)
	Debugf("Empty Hall Spots: %s", emptyHalls)
	Debugf("Empty Room Spots: %s", emptyRooms)
	for _, amph := range movableAmphs {
		// Check rooms first. If there's a valid, worthwhile room to room move, don't even check the halls for moves.
		roomMoveFound := false
		for _, room := range emptyRooms {
			if amph.IsHall || amph.Ind != room.Ind {
				if checkAndAppend(&Move{amph, room}) {
					roomMoveFound = true
					break
				}
			}
		}
		if roomMoveFound {
			continue
		}
		// No room moves, add all possible hall moves.
		if !amph.IsHall {
			for _, hall := range emptyHalls {
				checkAndAppend(&Move{amph, hall})
			}
		}
	}
	return rv
}

type MoveNodeList []*MoveNode

func (l MoveNodeList) String() string {
	lines := make([]string, len(l))
	for i, m := range l {
		lines[i] = m.String()
	}
	return strings.Join(AddLineNumbers(lines, 0), "\n")
}

type MoveNode struct {
	Move
	Cost  int     // the cost of making this move.
	State *Burrow // the state AFTER this move is applied.
}

func (m MoveNode) String() string {
	return fmt.Sprintf("From: %s, To: %s, Cost: %d", m.From, m.To, m.Cost)
}

func NewMoveNode(stateBeforeMove *Burrow, move *Move) (rv *MoveNode, err error) {
	rv = &MoveNode{
		Move:  *move,
		State: stateBeforeMove.CopyOf(),
	}
	rv.Cost, err = rv.State.MakeMove(rv.From, rv.To)
	return rv, err
}

type Move struct {
	From Spot
	To   Spot
}

func (m Move) String() string {
	return fmt.Sprintf("From: %s, To: %s", m.From, m.To)
}

func RoomRoomMove(r1, d1, r2, d2 int) *Move {
	return &Move{
		From: RoomSpot(r1, d1),
		To:   RoomSpot(r2, d2),
	}
}

func RoomHall(r, d, h int) *Move {
	return &Move{
		From: RoomSpot(r, d),
		To:   HallSpot(h),
	}
}

func HallRoom(h, r, d int) *Move {
	return &Move{
		From: HallSpot(h),
		To:   RoomSpot(r, d),
	}
}

func BurrowMustMove(b *Burrow, from, to Spot) int {
	rv, err := b.MakeMove(from, to)
	if err != nil {
		Stderr("Burrow:\n%s", b)
		panic(err)
	}
	Debugf("After move from %s to %s, Cost: %d\n%s", from, to, rv, b)
	return rv
}

func (b *Burrow) GetMoveCost(from, to Spot) (int, error) {
	amph := b.Get(from)
	if amph == '.' {
		return 0, fmt.Errorf("cannot move from %s to %s: does not have an amphipod.", from, to)
	}
	if toB := b.Get(to); toB != '.' {
		return 0, fmt.Errorf("cannot move from %s to %s: it is already occupied by %c.", from, to, toB)
	}
	// If coming from a room, make sure there's nothing above it.
	if !from.IsHall {
		// If moving to a spot in the same hall, we can skip a bunch of stuff.
		if !to.IsHall && from.Ind == to.Ind {
			f := from.Ind
			t := to.Ind
			if f > t {
				f, t = t, f
			}
			d := t - f
			for i := f + 1; i < t; i++ {
				if b.Rooms[from.Ind][f] != '.' {
					return 0, fmt.Errorf("cannot move from %s to %s: there's something in the room in the way.", from, to)
				}
			}
			return Cost[amph] * d, nil
		}
		// Either moving to the hall or a different room. either way, make sure they can get out.
		for i := from.Ind2 - 1; i >= 0; i-- {
			if b.Rooms[from.Ind][i] != '.' {
				return 0, fmt.Errorf("cannot move from %s to %s: there's something in the way above it.", from, to)
			}
		}
	}
	// If moving to a room, make sure it's the right room and there's nothing in the way.
	if !to.IsHall {
		// Make sure it's the right room.
		if to.Ind != AmphHome[amph] {
			return 0, fmt.Errorf("cannot move from %s to %s: %c cannot enter room %d", from, to, amph, to.Ind)
		}
		// Make sure it's either empty or has only the same type of amph.
		for i := 0; i < 4; i++ {
			if b.Rooms[to.Ind][i] != '.' && b.Rooms[to.Ind][i] != amph {
				return 0, fmt.Errorf("cannot move from %s to %s: room is not yet empty.", from, to)
			}
		}
		// Make sure it's empty up to the destination.
		for i := 0; i < to.Ind2; i++ {
			if b.Rooms[to.Ind][i] != '.' {
				return 0, fmt.Errorf("cannot move from %s to %s: there's a %c in the way at Room[%d][%d].", from, to, b.Rooms[to.Ind][i], to.Ind, i)
			}
		}
	}
	// Figure out the hall cells to check.
	// Room 0 empties into either Hall 1 or Hall 2.
	// Room 1 empties into either Hall 2 or Hall 3.
	// Room 2 empties into either Hall 3 or Hall 4.
	// Room 3 empties into either Hall 4 or Hall 5.
	// Need to check from the room exit to the destination, but the room exit is different depending on the direction.
	// Each hall spot checked will count as 2 steps (with some adjustments for other factors).
	var fromi, toi, di, steps int
	switch {
	case from.IsHall && to.IsHall:
		return 0, fmt.Errorf("cannot move from %s to %s: illegal hall-to-hall move.", from, to)
	case from.IsHall && !to.IsHall:
		if from.Ind <= to.Ind+1 {
			fromi = from.Ind + 1
			toi = to.Ind + 1
			di = 1
		} else {
			fromi = from.Ind - 1
			toi = to.Ind + 2
			di = -1
		}
		// Need to count the two steps it takes to get into the first room spot.
		steps += 2
		// Debugf("hr Steps: %d", steps)
	case !from.IsHall && to.IsHall:
		if from.Ind+1 < to.Ind {
			fromi = from.Ind + 2
			toi = to.Ind
			di = 1
		} else {
			fromi = from.Ind + 1
			toi = to.Ind
			di = -1
		}
		// No extra steps to count.
		// Debugf("rh Steps: %d", steps)
	case !from.IsHall && !to.IsHall:
		if from.Ind < to.Ind {
			fromi = from.Ind + 2
			toi = to.Ind + 1
			di = 1
		} else {
			fromi = from.Ind + 1
			toi = to.Ind + 2
			di = -1
		}
		// Again, need to count the two steps it takes to get into the first room spot.
		steps += 2
		// Debugf("rr Steps: %d", steps)
	}
	test := func(i int) bool {
		return i <= toi
	}
	if di < 0 {
		test = func(i int) bool {
			return i >= toi
		}
	}
	// Debugf("Hall Test Params: fromi: %d, toi: %d, di: %d", fromi, toi, di)
	// Check all the hall spots, and count the steps.
	for i := fromi; test(i); i += di {
		if b.Hall[i] != '.' {
			return 0, fmt.Errorf("cannot move from %s to %s: there's a %c in the way at Hall[%d].", from, to, b.Hall[i], i)
		}
		steps += 2
		// Debugf("hc Steps: %d", steps)
	}
	// Remove a step if coming from one of the hall ends (it was counted as 2 earlier).
	if from.IsHall && (from.Ind == 0 || from.Ind == 6) {
		steps--
		// Debugf("hf Steps: %d", steps)
	}
	// Remove a step if going to one of the hall ends (it was counted as 2 earlier).
	if to.IsHall && (to.Ind == 0 || to.Ind == 6) {
		steps--
		// Debugf("ht Steps: %d", steps)
	}
	// If coming from a room, count steps for each cell to get to the room's entry.
	if !from.IsHall {
		steps += from.Ind2
		// Debugf("rf Steps: %d", steps)
	}
	// If going to a room, count steps for each cell to get from the room's entry to the destination.
	if !to.IsHall {
		steps += to.Ind2
		// Debugf("rt Steps: %d", steps)
	}
	return Cost[amph] * steps, nil
}

func (b *Burrow) MakeMove(from, to Spot) (int, error) {
	cost, err := b.GetMoveCost(from, to)
	if err != nil {
		return 0, err
	}
	b.Set(to, b.Get(from))
	b.Set(from, '.')
	return cost, nil
}

type Spot struct {
	IsHall bool
	Ind    int
	Ind2   int
}

func (s Spot) String() string {
	if s.IsHall {
		return fmt.Sprintf("Hall[%d]", s.Ind)
	}
	return fmt.Sprintf("Room[%d][%d]", s.Ind, s.Ind2)
}

func HallSpot(x int) Spot {
	return Spot{
		IsHall: true,
		Ind:    x,
	}
}

func RoomSpot(room, depth int) Spot {
	return Spot{
		Ind:  room,
		Ind2: depth,
	}
}

func (s Spot) Equals(t Spot) bool {
	return s.IsHall == t.IsHall && s.Ind == t.Ind && s.Ind2 == t.Ind2
}

type Burrow struct {
	Hall  []byte
	Rooms [][]byte
}

func (b Burrow) String() string {
	lines := []string{
		"#############",
		fmt.Sprintf("#%c%c.%c.%c.%c.%c%c#", b.Hall[0], b.Hall[1], b.Hall[2], b.Hall[3], b.Hall[4], b.Hall[5], b.Hall[6]),
		fmt.Sprintf("###%c#%c#%c#%c###", b.Rooms[0][0], b.Rooms[1][0], b.Rooms[2][0], b.Rooms[3][0]),
	}
	for i := 1; i < 4; i++ {
		lines = append(lines, fmt.Sprintf("  #%c#%c#%c#%c#  ", b.Rooms[0][i], b.Rooms[1][i], b.Rooms[2][i], b.Rooms[3][i]))
	}
	lines = append(lines, "  #########")
	return strings.Join(lines, "\n")
}

func NewBurrow() *Burrow {
	return &Burrow{
		Hall: []byte("......."),
		Rooms: [][]byte{
			[]byte("...."),
			[]byte("...."),
			[]byte("...."),
			[]byte("...."),
		},
	}
}

func (b Burrow) CopyOf() *Burrow {
	rv := Burrow{
		Hall:  make([]byte, len(b.Hall)),
		Rooms: make([][]byte, len(b.Rooms)),
	}
	for i, b := range b.Hall {
		rv.Hall[i] = b
	}
	for i, r := range b.Rooms {
		rv.Rooms[i] = make([]byte, len(r))
		for j, b := range r {
			rv.Rooms[i][j] = b
		}
	}
	return &rv
}

func (b Burrow) Get(s Spot) byte {
	if s.IsHall {
		return b.Hall[s.Ind]
	}
	return b.Rooms[s.Ind][s.Ind2]
}

func (b *Burrow) Set(s Spot, a byte) {
	if s.IsHall {
		b.Hall[s.Ind] = a
	} else {
		b.Rooms[s.Ind][s.Ind2] = a
	}
}

func (b Burrow) Equals(c *Burrow) bool {
	if !bytes.Equal(b.Hall, c.Hall) {
		return false
	}
	if len(b.Rooms) != len(c.Rooms) {
		return false
	}
	for i := range b.Rooms {
		if !bytes.Equal(b.Rooms[i], c.Rooms[i]) {
			return false
		}
	}
	return true
}

type Input struct {
	Start *Burrow
}

func (i Input) String() string {
	return fmt.Sprintf("Burrow:\n%s", i.Start)
}

func ParseInput(lines []string) (*Input, error) {
	rv := Input{}
	rv.Start = NewBurrow()
	rv.Start.Hall[0] = byte(lines[1][1])
	rv.Start.Hall[6] = byte(lines[1][11])
	for i := 1; i <= 5; i++ {
		rv.Start.Hall[i] = lines[1][i*2]
	}
	for r := 0; r < len(rv.Start.Rooms); r++ {
		for d := 0; d < len(rv.Start.Rooms[r]); d++ {
			rv.Start.Rooms[r][d] = lines[2+d][3+r*2]
		}
	}
	return &rv, nil
}

// -------------------------------------------------------------------------------------
// -------------------------------  Some generic stuff  --------------------------------
// -------------------------------------------------------------------------------------

// PQNode is a node for use in a priority queue.
type PQNode struct {
	Ele      *MoveNode
	MinCost  int
	Index    int
	Visited  bool
	ComeFrom *PQNode
}

func (n PQNode) String() string {
	return fmt.Sprintf("%d: MinCost: %d, Visited: %t, Ele: %s", n.Index, n.MinCost, n.Visited, n.Ele)
}

// PriorityQueue implements heap.Interface and holds PQNodes.
type PriorityQueue []*PQNode

// Len returns the length of this priority queue.
func (pq PriorityQueue) Len() int { return len(pq) }

// Less returns true if entry i is less than entry j.
func (pq PriorityQueue) Less(i, j int) bool {
	// If risks are different, use that for comparisons first.
	if pq[i].MinCost != pq[j].MinCost {
		return pq[i].MinCost < pq[j].MinCost
	}
	// Use the min move cost next.
	if pq[i].Ele.Cost != pq[j].Ele.Cost {
		return pq[i].Ele.Cost < pq[j].Ele.Cost
	}
	// Really shouldn't have gotten past that last one, so use the indexes.
	if pq[i].Index != pq[j].Index {
		return pq[i].Index < pq[j].Index
	}
	// Seriously, wtf.
	return i < j
}

// Swap swaps the entries at i and j.
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

// Push adds an node to the priority queue.
func (pq *PriorityQueue) Push(x interface{}) {
	node := x.(*PQNode)
	node.Index = len(*pq)
	*pq = append(*pq, node)
}

// Pop removes the next node from the priority queue.
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	old[n-1] = nil  // avoid memory leak
	node.Index = -1 // for safety
	node.Visited = true
	*pq = old[0 : n-1]
	return node
}

// update modifies the risk of a PQNode in the priority queue.
func (pq *PriorityQueue) Update(node *PQNode, minCost int, comeFrom *PQNode) {
	node.MinCost = minCost
	node.ComeFrom = comeFrom
	heap.Fix(pq, node.Index)
}

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
		if arg[0] == '-' {
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
