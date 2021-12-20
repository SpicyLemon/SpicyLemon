package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if err := Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func Run() error {
	if len(os.Args) == 1 {
		fmt.Println("Usage: go run transform-point.go 'x,y,z' ['x,y,z' ...]")
		return nil
	}
	for i, arg := range os.Args[1:] {
		if i != 0 {
			fmt.Println()
		}
		fmt.Printf("Input %d: %q\n", i, arg)
		pt, err := ParsePoint(arg)
		if err != nil {
			return err
		}
		r := pt.GetAllRotations()
		v := pt.GetAllVersions()
		vstr := []byte(v.String())
		blank := bytes.Repeat([]byte(" "), 16)
		for _, p := range r {
			i := bytes.Index(vstr, []byte(p.String()))
			copy(vstr[i:], blank)
		}
		if len(r.GetDuplicates()) != 0 {
			fmt.Printf("%d: Point %s has duplicate rotations!\n", i, pt)
		}
		if len(v.GetDuplicates()) != 0 {
			fmt.Printf("%d: Point %s has duplicate versions!\n", i, pt)
		}
		fmt.Printf("%d: Point %s in all rotations (%d):\n%s", i, pt, len(r), r)
		fmt.Printf("%d: Point %s in all versions (%d):\n%s", i, pt, len(v), v)
		fmt.Printf("%d: Point %s in versions but not rotations (%d):\n%s", i, pt, len(v)-len(r), vstr)
	}
	return nil
}

func (l PointList) GetDuplicates() PointList {
	rv := PointList{}
	for i := 0; i < len(l)-1; i++ {
		for j := i + 1; j < len(l); j++ {
			if !rv.Contains(l[j]) && l[i].Equals(l[j]) {
				rv = append(rv, l[j])
			}
		}
	}
	return rv
}

func (l PointList) Contains(pt *Point) bool {
	for _, p := range l {
		if p.Equals(pt) {
			return true
		}
	}
	return false
}

func AppendIfNew(l PointList, pt *Point) PointList {
	if !l.Contains(pt) {
		l = append(l, pt)
	}
	return l
}

var pm = []int{1, -1}

func (p Point) GetAllRotations() PointList {
	x, y, z := p.X, p.Y, p.Z
	orientations := []*Point{
		NewPoint(x, y, z),
		NewPoint(y, -x, z),
		NewPoint(-x, -y, z),
		NewPoint(-y, x, z),
		NewPoint(z, y, -x),
		NewPoint(-z, y, x),
	}
	rv := PointList{}
	for _, o := range orientations {
		rv = AppendIfNew(rv, o)
		rv = AppendIfNew(rv, NewPoint(o.X, -o.Z, o.Y))
		rv = AppendIfNew(rv, NewPoint(o.X, -o.Y, -o.Z))
		rv = AppendIfNew(rv, NewPoint(o.X, o.Z, -o.Y))
	}
	sort.Sort(rv)
	return rv
}

func (p Point) GetAllVersions() PointList {
	x, y, z := p.X, p.Y, p.Z
	pts := PointList{
		&Point{x, y, z},
		&Point{x, z, y},
		&Point{y, x, z},
		&Point{y, z, x},
		&Point{z, x, y},
		&Point{z, y, x},
	}
	rv := PointList{}
	for _, p := range pts {
		for _, dx := range pm {
			for _, dy := range pm {
				for _, dz := range pm {
					rv = AppendIfNew(rv, &Point{p.X * dx, p.Y * dy, p.Z * dz})
				}
			}
		}
	}
	sort.Sort(rv)
	return rv
}

type Point struct {
	X int
	Y int
	Z int
}

func (p Point) String() string {
	return fmt.Sprintf("(% 4d,% 4d,% 4d)", p.X, p.Y, p.Z)
}

func NewPoint(x, y, z int) *Point {
	return &Point{
		X: x,
		Y: y,
		Z: z,
	}
}

func ParsePoint(str string) (*Point, error) {
	parts := strings.Split(str, ",")
	rv := Point{}
	var err error
	if len(parts) < 2 {
		return nil, fmt.Errorf("could not parse %q to Point: invalid format", str)
	}
	rv.X, err = strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse %q to Point: %w", str, err)
	}
	rv.Y, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not parse %q to Point: %w", str, err)
	}
	if len(parts) > 2 {
		rv.Z, err = strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("could not parse %q to Point: %w", str, err)
		}
	}
	return &rv, nil
}

func (p Point) Equals(pt *Point) bool {
	return p.X == pt.X && p.Y == pt.Y && p.Z == pt.Z
}

type PointList []*Point

func (l PointList) String() string {
	leadFmt := "  " + DigitFormatForMax(len(l)) + ":"
	lastI := len(l) - 1
	var rv strings.Builder
	for i, p := range l {
		if i%8 == 0 {
			rv.WriteString(fmt.Sprintf(leadFmt, i))
		}
		rv.WriteByte(' ')
		rv.WriteString(p.String())
		if i != lastI && i%8 == 7 {
			rv.WriteByte('\n')
		}
	}
	rv.WriteByte('\n')
	return rv.String()
}

func (p PointList) Len() int      { return len(p) }
func (p PointList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PointList) Less(i, j int) bool {
	if p[i].Z != p[j].Z {
		return p[i].Z < p[j].Z
	}
	if p[i].Y != p[j].Y {
		return p[i].Y < p[j].Y
	}
	return p[i].X < p[j].X
}

func DigitFormatForMax(max int) string {
	return fmt.Sprintf("%%%dd", len(fmt.Sprintf("%d", max)))
}
