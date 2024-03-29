package main

import (
	"container/heap"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	Debugf("Parsed Input:\n%s", input)
	var graph *Graph
	if params.HasCustom("cut") {
		graph, err = CustomCut(input.Graph, params)
	} else {
		graph = FindAndCut(input.Graph, params)
	}
	if err != nil {
		return "", err
	}
	Debugf("After Cut:\n%s", graph)
	counts := CountGroups(graph, params)
	Debugf("Group counts: %v", counts)
	answer := MulInts(counts)
	return fmt.Sprintf("%d", answer), nil
}

func FindAndCut(graph *Graph, params *Params) *Graph {
	cutsLeft := 3
	for cutsLeft > 0 {
		cuts := FindCuts(graph, cutsLeft, params)
		Debugf("Cuts: %s", SliceToStrings(cuts))
		if len(cuts) == 0 {
			return graph
		}
		for _, cut := range cuts {
			params.Verbosef("Making cut: %s", cut)
			graph.CutConnection(cut)
			cutsLeft--
		}
	}
	return graph
}

func FindCuts(graph *Graph, cutsLeft int, params *Params) []*GraphEdge {
	minMax := 4
	if params.InputFile != DEFAULT_INPUT_FILE {
		minMax = 6
	}
	names := MapSlice(graph.GetNodes(), (*GraphNode).GetName)
	var paths [][]*GraphEdge
	i := 0
	di := 1
	for {
		j := (i + di) % len(names)
		path := FindPath(graph, names[i], names[j], params)
		if path == nil {
			panic(fmt.Errorf("no path found from %s to %s", names[i], names[j]))
		}
		if len(path) >= 3 {
			edges := PathToEdges(path)
			if len(edges) != len(path)-1 {
				panic(fmt.Errorf("could not convert path [%s] to edges [%s]",
					strings.Join(SliceToStrings(path), " : "),
					strings.Join(SliceToStrings(edges), " : ")))
			}
			paths = append(paths, edges)
		}
		if len(paths) >= 4 {
			rv := GetMostCommonEdges(paths, minMax)
			if debug {
				Stderrf("Paths: %d, Most Common Edges (%d): %s",
					len(paths), len(rv), strings.Join(SliceToStrings(rv), ", "))
			}
			if len(rv) > 0 && len(rv) <= cutsLeft {
				return rv
			}
		}
		i++
		if i >= len(names) {
			i = 0
			di++
		}
	}
}

func GetMostCommonEdges(paths [][]*GraphEdge, minMax int) []*GraphEdge {
	keyer := func(edge *GraphEdge) string {
		a, b := OrderStrings(edge.A.Name, edge.B.Name)
		return fmt.Sprintf("%s-%s", a, b)
	}

	counts := make(map[string]int)
	edgeMap := make(map[string]*GraphEdge)
	for _, path := range paths {
		// if debug {
		//	Stderrf("Counting edges: %s", SliceToStrings(path))
		// }
		for _, edge := range path {
			// Debugf("Counting: %s", edge)
			key := keyer(edge)
			counts[key]++
			edgeMap[key] = edge
		}
	}

	Debugf("Counts: %v", counts)

	var max int
	var maxes []*GraphEdge
	for key, count := range counts {
		switch {
		case count > max:
			max = count
			maxes = []*GraphEdge{edgeMap[key]}
		case count == max:
			maxes = append(maxes, edgeMap[key])
		}
	}

	if max < minMax {
		Debugf("Max Count %d is less than %d", max, minMax)
		return nil
	}

	return maxes
}

func PathToEdges(path []*GraphNode) []*GraphEdge {
	var rv []*GraphEdge
	for i := 0; i < len(path)-1; i++ {
		edge := path[i].GetEdge(path[i+1].Name)
		if edge == nil {
			panic(fmt.Errorf("path has no edge from [%d]: %s to [%d] %s", i, path[i], i+1, path[i+1].Name))
		}
		rv = append(rv, edge)
	}
	return rv
}

func MulInts(vals []int) int {
	rv := 1
	for _, v := range vals {
		rv *= v
	}
	return rv
}

func FindPath(graph *Graph, from, to string, params *Params) []*GraphNode {
	// defer FuncEnding(FuncStarting())
	finder := NewPathFinder(graph, from, to)
	finder.FindSolution(params)
	if debug {
		Debugf("Path from %s to %s: %s", from, to, strings.Join(MapSlice(finder.Solution, (*GraphNode).GetName), " "))
	}
	return finder.Solution
}

type PathFinder struct {
	Graph     *Graph
	From      string
	To        string
	Unvisited PriorityQueue
	Visited   map[string]*PathNode
	Last      *PathNode
	Solution  []*GraphNode
}

func NewPathFinder(graph *Graph, from, to string) *PathFinder {
	node := graph.Get(from)
	if node == nil {
		panic(fmt.Errorf("node %q does not exist", from))
	}
	rv := &PathFinder{
		Graph:     graph,
		From:      from,
		To:        to,
		Unvisited: make(PriorityQueue, 0, 100),
		Visited:   make(map[string]*PathNode),
	}
	rv.Unvisited.Push(NewPathNode(node, nil))
	// Debugf("Initializing Solver: %s", rv)
	heap.Init(&rv.Unvisited)
	return rv
}

func (f PathFinder) String() string {
	solution := "<nil>"
	if f.Solution != nil {
		solution = fmt.Sprintf("(%d):\n%s", len(f.Solution), StringNumberJoin(f.Solution, 0, "\n"))
	}
	return fmt.Sprintf("%s to %s:\nUnvisited: %d\nVisited: %d\nLast: %s\nSolution: %s",
		f.From, f.To, len(f.Unvisited), len(f.Visited), f.Last, solution)
}

func (f PathFinder) IsDone() bool {
	return f.Solution != nil || len(f.Unvisited) == 0
}

func (f *PathFinder) FindSolution(params *Params) {
	// defer FuncEnding(FuncStarting())
	i := 0
	for !f.IsDone() {
		i++
		f.CalculateNext()
		if params.Verbose && i%1_000_000 == 0 {
			Stderrf("Calculated %d million moves.\nPathFinder:%s", i/1_000_000, f)
		}
	}
}

func (f *PathFinder) CalculateNext() {
	keyNode := heap.Pop(&f.Unvisited).(*PathNode) //nolint:forcetypeassert // want panic here.
	// Debugf("Key Node: %s", keyNode)
	// Debugf("Path Finder: %s", f)
	f.Last = keyNode
	f.Visited[keyNode.Name] = keyNode

	keyEdges := keyNode.GetEdges()
	for _, edge := range keyEdges {
		if edge.A.Name == f.To {
			f.Solution = MapSlice(keyNode.GetPath(), (*PathNode).GetGraphNode)
			f.Solution = append(f.Solution, edge.A)
			return
		}
		if edge.B.Name == f.To {
			f.Solution = MapSlice(keyNode.GetPath(), (*PathNode).GetGraphNode)
			f.Solution = append(f.Solution, edge.B)
			return
		}
		if f.Visited[edge.A.Name] == nil {
			pn := NewPathNode(edge.A, keyNode)
			heap.Push(&f.Unvisited, pn)
		}
		if f.Visited[edge.B.Name] == nil {
			pn := NewPathNode(edge.B, keyNode)
			heap.Push(&f.Unvisited, pn)
		}
	}
}

type PriorityQueue []*PathNode

func (q PriorityQueue) String() string {
	return strings.Join(SliceToStrings(q), "\n")
}

func (q PriorityQueue) Len() int {
	return len(q)
}

func (q PriorityQueue) Less(i, j int) bool {
	if q[i].Length != q[j].Length {
		return q[i].Length < q[j].Length
	}
	return q[i].Name < q[j].Name
}

func (q PriorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].Index = i
	q[j].Index = j
}

func (q *PriorityQueue) Push(x interface{}) {
	node := x.(*PathNode) //nolint:forcetypeassert // want panic here.
	node.Index = len(*q)
	*q = append(*q, node)
}

func (q *PriorityQueue) Pop() interface{} {
	old := *q
	n := len(old)
	node := old[n-1]
	old[n-1] = nil  // avoid memory leak
	node.Index = -1 // for safety
	*q = old[0 : n-1]
	return node
}

func (q *PriorityQueue) Update(node *PathNode, prev *PathNode) {
	node.Length = prev.Length + 1
	node.Prev = prev
	heap.Fix(q, node.Index)
}

type PathNode struct {
	GraphNode
	Prev   *PathNode
	Length int
	Index  int
}

func NewPathNode(node *GraphNode, prev *PathNode) *PathNode {
	rv := &PathNode{GraphNode: *node, Length: 1, Index: -1}
	if prev != nil {
		rv.Prev = prev
		rv.Length = prev.Length + 1
	}
	return rv
}

func (p PathNode) String() string {
	prev := "<nil>"
	if p.Prev != nil {
		prev = p.Prev.Name
	}
	return fmt.Sprintf("[%d]: (%d) %s From %s", p.Index, p.Length, p.GraphNode, prev)
}

func (p PathNode) GetLength() int {
	return p.Length
}

func (p PathNode) GetPath() []*PathNode {
	rv := make([]*PathNode, 1, p.Length)
	cur := &p
	rv[0] = cur
	for cur.Prev != nil {
		cur = cur.Prev
		rv = append(rv, cur)
	}
	slices.Reverse(rv)
	return rv
}

func (p PathNode) GetGraphNode() *GraphNode {
	return &p.GraphNode
}

type NodeGroup map[string]*GraphNode

func NewNodeGroup() NodeGroup {
	return make(NodeGroup)
}

func (g NodeGroup) String() string {
	return strings.Join(StringKeys(g), " ")
}

func (g NodeGroup) Add(node *GraphNode) {
	g[node.Name] = node
}

func (g NodeGroup) AddAll(group NodeGroup) {
	for k, v := range group {
		g[k] = v
	}
}

func (g NodeGroup) Has(node *GraphNode) bool {
	return g[node.Name] != nil
}

func (g NodeGroup) HasOneOf(nodes []*GraphNode) bool {
	for _, node := range nodes {
		if g.Has(node) {
			return true
		}
	}
	return false
}

func (g NodeGroup) GetNodes() []*GraphNode {
	return OrderedVals(g)
}

func (g NodeGroup) Count() int {
	return len(g)
}

func CountGroups(graph *Graph, params *Params) []int {
	groupMap := make(map[string]NodeGroup)
	var groups []NodeGroup
	allNodes := graph.GetNodes()
	for _, node := range allNodes {
		var nodeGroups []NodeGroup
		for _, edge := range node.GetEdges() {
			if group := groupMap[edge.A.Name]; group != nil {
				nodeGroups = append(nodeGroups, group)
			}
			if group := groupMap[edge.B.Name]; group != nil {
				nodeGroups = append(nodeGroups, group)
			}
		}
		var group NodeGroup
		switch len(nodeGroups) {
		case 0:
			group = NewNodeGroup()
			groups = append(groups, group)
		case 1:
			group = nodeGroups[0]
		default:
			group = NewNodeGroup()
			for _, subGroup := range nodeGroups {
				for name, subNode := range subGroup {
					group.Add(subNode)
					groupMap[name] = group
				}
			}
			var newGroups []NodeGroup
			newGroupNodes := group.GetNodes()
			for _, oldGroup := range groups {
				if !oldGroup.HasOneOf(newGroupNodes) {
					newGroups = append(newGroups, oldGroup)
				}
			}
			newGroups = append(newGroups, group)
			groups = newGroups
		}
		group.Add(node)
		groupMap[node.Name] = group
	}

	if params.Verbose {
		Stdoutf("Node Groups (%d):\n%s", len(groups), StringNumberJoin(groups, 1, "\n"))
	}

	return MapSlice(groups, NodeGroup.Count)
}

func CombineGroups(groups []NodeGroup) NodeGroup {
	rv := NewNodeGroup()
	for _, group := range groups {
		rv.AddAll(group)
	}
	return rv
}

func CustomCut(graph *Graph, params *Params) (*Graph, error) {
	cuts := make([]*GraphEdge, 0, len(params.Custom)-1)
	inCut := false
	for _, custom := range params.Custom {
		if custom == "cut" {
			inCut = true
			continue
		}
		if !inCut {
			continue
		}
		parts := strings.Split(custom, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid custom cut pair: %q", custom)
		}
		node1 := graph.Get(parts[0])
		if node1 == nil {
			return nil, fmt.Errorf("no such node: %q", parts[0])
		}
		node2 := graph.Get(parts[1])
		if node2 == nil {
			return nil, fmt.Errorf("no such node: %q", parts[1])
		}
		cuts = append(cuts, NewGraphEdge(node1, node2))
	}

	for _, cut := range cuts {
		graph.CutConnection(cut)
	}

	return graph, nil
}

type Graph struct {
	Nodes map[string]*GraphNode
	Edges map[string]map[string]*GraphEdge
}

func NewGraph() *Graph {
	return &Graph{}
}

func (g *Graph) AddConnection(name1, name2 string) {
	node1 := g.Get(name1)
	if node1 == nil {
		node1 = NewGraphNode(name1)
		g.Add(node1)
	}
	node2 := g.Get(name2)
	if node2 == nil {
		node2 = NewGraphNode(name2)
		g.Add(node2)
	}
	edge := g.GetEdge(name1, name2)
	if edge == nil {
		edge = node1.GetEdge(name2)
	}
	if edge == nil {
		edge = node2.GetEdge(name1)
	}
	if edge == nil {
		edge = NewGraphEdge(node1, node2)
	}
	g.AddEdge(edge)
	node1.AddEdge(edge)
	node2.AddEdge(edge)
}

func (g *Graph) CutConnection(edge *GraphEdge) {
	edge.A.RemoveEdge(edge.B.Name)
	edge.B.RemoveEdge(edge.A.Name)
	g.RemoveEdge(edge.A.Name, edge.B.Name)
}

func (g Graph) Has(name string) bool {
	return g.Nodes != nil && g.Nodes[name] != nil
}

func (g Graph) Get(name string) *GraphNode {
	if g.Nodes == nil {
		return nil
	}
	return g.Nodes[name]
}

func (g Graph) GetEdge(name1, name2 string) *GraphEdge {
	if g.Edges == nil {
		return nil
	}
	var rv *GraphEdge
	if g.Edges[name1] != nil {
		rv = g.Edges[name1][name2]
	}
	if rv == nil && g.Edges[name2] != nil {
		rv = g.Edges[name2][name1]
	}
	return rv
}

func (g *Graph) Add(node *GraphNode) {
	if g.Nodes == nil {
		g.Nodes = make(map[string]*GraphNode)
	}
	g.Nodes[node.Name] = node
	for _, edge := range node.GetEdges() {
		g.AddEdge(edge)
	}
}

func (g *Graph) AddEdge(edge *GraphEdge) {
	if g.Edges == nil {
		g.Edges = make(map[string]map[string]*GraphEdge)
	}
	nameA := edge.A.Name
	nameB := edge.B.Name
	if g.Edges[nameA] == nil {
		g.Edges[nameA] = make(map[string]*GraphEdge)
	}
	if g.Edges[nameB] == nil {
		g.Edges[nameB] = make(map[string]*GraphEdge)
	}
	g.Edges[nameA][nameB] = edge
	g.Edges[nameB][nameA] = edge
}

func (g *Graph) Remove(name string) {
	if g.Nodes == nil {
		return
	}
	node := g.Nodes[name]
	if node == nil {
		return
	}
	delete(g.Nodes, name)
	if g.Edges != nil {
		for _, edge := range node.GetEdges() {
			g.RemoveEdge(edge.A.Name, edge.B.Name)
		}
	}
}

func (g *Graph) RemoveEdge(a, b string) {
	if g.Edges == nil {
		return
	}
	delete(g.Edges[a], b)
	delete(g.Edges[b], a)
	if len(g.Edges[a]) == 0 {
		delete(g.Edges, a)
	}
	if len(g.Edges[b]) == 0 {
		delete(g.Edges, b)
	}
}

func (g Graph) GetNodes() []*GraphNode {
	return OrderedVals(g.Nodes)
}

func (g Graph) GetEdges() []*GraphEdge {
	var rv []*GraphEdge
	for _, a := range StringKeys(g.Edges) {
		for _, b := range StringKeys(g.Edges[a]) {
			if b <= a {
				continue
			}
			rv = append(rv, g.Edges[a][b])
		}
	}
	return rv
}

func (g Graph) GetNodeEdges(name string) []*GraphEdge {
	if g.Edges == nil {
		return nil
	}
	return OrderedVals(g.Edges[name])
}

func (g Graph) String() string {
	nodes := g.GetNodes()
	edges := g.GetEdges()
	return fmt.Sprintf("Nodes (%d):\n%s\nEdges (%d):\n%s",
		len(nodes), StringNumberJoin(nodes, 1, "\n"),
		len(edges), StringNumberJoin(edges, 1, "\n"),
	)
}

type GraphNode struct {
	Name  string
	Edges map[string]*GraphEdge
}

func NewGraphNode(name string) *GraphNode {
	return &GraphNode{Name: name}
}

func (n GraphNode) String() string {
	return fmt.Sprintf("%s: %s", n.Name, strings.Join(StringKeys(n.Edges), " "))
}

func (n GraphNode) GetName() string {
	return n.Name
}

func (n *GraphNode) AddEdge(edge *GraphEdge) {
	if n.Edges == nil {
		n.Edges = make(map[string]*GraphEdge)
	}
	switch n.Name {
	case edge.A.Name:
		n.Edges[edge.B.Name] = edge
	case edge.B.Name:
		n.Edges[edge.A.Name] = edge
	default:
		panic(fmt.Errorf("cannot add edge %s to node %s", edge, n.Name))
	}
}

func (n *GraphNode) RemoveEdge(name string) {
	delete(n.Edges, name)
}

func (n *GraphNode) GetEdge(name string) *GraphEdge {
	if n.Edges == nil {
		return nil
	}
	return n.Edges[name]
}

func (n *GraphNode) Equal(n2 *GraphNode) bool {
	if n == n2 {
		return true
	}
	return n.Name == n2.Name
}

func (n GraphNode) GetEdges() []*GraphEdge {
	return OrderedVals(n.Edges)
}

type GraphEdge struct {
	A *GraphNode
	B *GraphNode
}

func NewGraphEdge(n1, n2 *GraphNode) *GraphEdge {
	if n1.Name < n2.Name {
		return &GraphEdge{A: n1, B: n2}
	}
	if n1.Name > n2.Name {
		return &GraphEdge{A: n2, B: n1}
	}
	return nil
}

func (e GraphEdge) String() string {
	return fmt.Sprintf("%s -- %s", e.A.Name, e.B.Name)
}

func (e *GraphEdge) Equal(e2 *GraphEdge) bool {
	if e == e2 {
		return true
	}
	return e.A.Equal(e2.A) && e.B.Equal(e2.B)
}

func (e GraphEdge) GetNames() []string {
	return []string{e.A.Name, e.B.Name}
}

func OrderStrings(a, b string) (string, string) {
	if a < b {
		return a, b
	}
	return b, a
}

type Input struct {
	Graph *Graph
}

func (i Input) String() string {
	return fmt.Sprintf("Graph:\n%s", i.Graph)
}

func ParseInput(lines []string) (*Input, error) {
	defer FuncEnding(FuncStarting())
	rv := Input{Graph: NewGraph()}
	for _, line := range lines {
		name, cons, err := ParseLine(line)
		if err != nil {
			return nil, err
		}
		for _, con := range cons {
			rv.Graph.AddConnection(name, con)
		}
	}
	return &rv, nil
}

func ParseLine(line string) (string, []string, error) {
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("could not parse line %q", line)
	}
	return parts[0], strings.Split(strings.TrimSpace(parts[1]), " "), nil
}

func StringKeys[M ~map[string]E, E any](m M) []string {
	rv := make([]string, 0, len(m))
	for k := range m {
		rv = append(rv, k)
	}
	slices.Sort(rv)
	return rv
}

func OrderedVals[M ~map[string]E, E any](m M) []E {
	if m == nil {
		return nil
	}
	rv := make([]E, len(m))
	for i, name := range StringKeys(m) {
		rv[i] = m[name]
	}
	return rv
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

// AddIndexNumbers adds index numbers to each string.
func AddIndexNumbers(lines []string, startAt int) []string {
	if len(lines) == 0 {
		return []string{}
	}
	rv := make([]string, len(lines))
	for i, line := range lines {
		rv[i] = fmt.Sprintf("[%d]%s", i+startAt, line)
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
