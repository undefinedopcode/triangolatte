package triangolatte

import (
	"container/list"
	"errors"
	"sort"
)

// Set of int values.
type Set map[int]bool

func cyclic(i, n int) int {
	return (i%n + n) % n
}

// isReflex checks if angle created by points a, b and c is reflex.
//
// Angle equal to math.Pi is considered convex for practical reasons (it can be
// used just fine in the triangulation).
func isReflex(a, b, c Point) bool {
	return (b.X-a.X)*(c.Y-b.Y)-(c.X-b.X)*(b.Y-a.Y) < 0
}

// isInsideTriangle checks if given point P lays inside triangle [A, B, C].
// Points on the edges are assumed to be inside.
func isInsideTriangle(a, b, c, p Point) bool {
	return (c.X-p.X)*(a.Y-p.Y)-(a.X-p.X)*(c.Y-p.Y) >= 0 &&
		(a.X-p.X)*(b.Y-p.Y)-(b.X-p.X)*(a.Y-p.Y) >= 0 &&
		(b.X-p.X)*(c.Y-p.Y)-(c.X-p.X)*(b.Y-p.Y) >= 0
}

// isEar checks if given element is an ear of the polygon.
func isEar(p *Element) bool {
	a, b, c := p.Prev.Point, p.Point, p.Next.Point
	if isReflex(a, b, c) {
		return false
	}

	r := p.Next.Next
	for r != p.Prev {
		inside := isInsideTriangle(a, b, c, r.Point)
		reflex := isReflex(r.Prev.Point, r.Point, r.Next.Point)
		if inside && reflex {
			return false
		}
		r = r.Next
	}
	return true
}

// findK finds the edges that intersect with ray `M + t * (1, 0)`. Let `K` be
// the closest visible point to `M` on this ray.
func findK(m Point, outer []Point) (k Point, k1, k2 int, err error) {
	for i, j := len(outer)-1, 0; j < len(outer); i, j = j, j+1 {
		// Skip edges that does not have their first point below `M` and the second
		// one above.
		if outer[i].Y > m.Y || outer[j].Y < m.Y {
			continue
		}

		// Calculate simplified intersection of ray (1, 0) and [V_i, V_j] segment.
		v1 := m.Sub(outer[i])
		v2 := outer[j].Sub(outer[i])

		t1 := v2.Cross(v1) / v2.Y
		t2 := v1.Y / v2.Y

		if t1 >= 0.0 && t2 >= 0.0 && t2 <= 1.0 {
			// If there is no current `k` candidate or this one is closer.
			if t1-m.X < k.X {
				k = Point{X: t1 + m.X, Y: m.Y}
				k1, k2 = i, j
				return
			}
		} else {
			err = errors.New("cannot calculate intersection, problematic data")
			return
		}
	}
	return
}

func areAllOutside(m, k Point, pIndex int, outer []Point) bool {
	allOutside := true
	for i := range outer {
		// We have to skip M, K and P vertices. Since M is from the inner
		// polygon and K was proved to not match any vertex, the only one to
		// check is pIndex
		if i == pIndex {
			continue
		}

		if isInsideTriangle(m, k, outer[pIndex], outer[i]) {
			allOutside = false
		}
	}
	return allOutside
}

func findClosest(m, k Point, pIndex int, outer []Point) int {
	reflex := list.New()
	n := len(outer)
	for i := 0; i < n; i++ {
		notInside := !isInsideTriangle(m, k, outer[pIndex], outer[i])
		prev, next := cyclic(i-1, n), cyclic(i+1, n)
		notReflex := !isReflex(outer[prev], outer[i], outer[next])
		if notInside || notReflex {
			continue
		}
		reflex.PushBack(i)
	}
	var closest int
	var maxDist float64

	for r := reflex.Front(); r != nil; r = r.Next() {
		i := r.Value.(int)
		dist := outer[i].Distance2(outer[closest])
		if dist > maxDist {
			closest = i
			maxDist = dist
		}
	}
	return closest
}

func combinePolygons(outer, inner []Point) ([]Point, error) {
	xMax := 0.0
	mIndex := 0
	for i := 0; i < len(inner); i++ {
		if inner[i].X > xMax {
			xMax = inner[i].X
			mIndex = i
		}
	}

	m := inner[mIndex]

	var pIndex int
	visibleIndex := -1

	k, k1, k2, err := findK(m, outer)

	if err != nil {
		return nil, err
	}

	// If `K` is vertex of the outer polygon, `M` and `K` are mutually visible.
	for i := 0; i < len(outer); i++ {
		if outer[i] == k {
			visibleIndex = i
		}
	}

	// Otherwise, `K` is an interior point of the edge `[V_k_1, V_k_2]`. Find `P`
	// which is endpoint with greater x-value.
	if outer[k1].X > outer[k2].X {
		pIndex = k1
	} else {
		pIndex = k2
	}

	// Check with all vertices of the outer polygon to be outside of the
	// triangle `[M, K, P]`. If it is true, `M` and `P` are mutually visible.
	allOutside := areAllOutside(m, k, pIndex, outer)

	if visibleIndex < 0 && allOutside {
		visibleIndex = pIndex
	}

	// Otherwise at least one reflex vertex lies in `[M, K, P]`. Search for the
	// array of reflex vertices `R` that minimizes the angle between `(1, 0)` and
	// line segment `[M, R]`. If there is exactly one vertex in `R` then they are
	// mutually visible. If there are multiple such vertices, pick the one closest
	// to `M`.
	if visibleIndex < 0 {
		visibleIndex = findClosest(m, k, pIndex, outer)
	}

	if visibleIndex < 0 {
		return nil, errors.New("could not find visible vertex")
	}

	result := make([]Point, 0, len(outer)+len(inner)+2)
	result = append(result, outer[:visibleIndex+1]...)
	for i := 0; i < len(inner); i++ {
		result = append(result, inner[cyclic(mIndex+i, len(inner))])
	}
	result = append(result, inner[mIndex], outer[visibleIndex])
	result = append(result, outer[visibleIndex+1:]...)

	return result, nil
}

type byMaxX [][]Point

func (polygons byMaxX) Len() int {
	return len(polygons)
}

func (polygons byMaxX) Swap(i, j int) {
	polygons[i], polygons[j] = polygons[j], polygons[i]
}

func (polygons byMaxX) Less(i, j int) bool {
	xMax := 0.0

	for k := 0; k < len(polygons[i]); k++ {
		if polygons[i][k].X > xMax {
			xMax = polygons[i][k].X
		}
	}

	for k := 0; k < len(polygons[j]); k++ {
		if polygons[j][k].X > xMax {
			return false
		}
	}

	return true
}

// JoinHoles removes holes, joining them with the rest of the polygon.
// Provides pre-processing for Polygon. First element of the points array is the
// outer polygon, the rest of them are considered as holes to be removed.
func JoinHoles(points [][]Point) ([]Point, error) {
	if len(points) == 0 {
		return nil, errors.New("cannot process empty points array")
	}

	if len(points) == 1 {
		return points[0], nil
	}

	sort.Sort(byMaxX(points[1:]))

	current := points[0]
	var err error

	for i := 1; i < len(points); i++ {
		current, err = combinePolygons(current, points[i])

		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

// Polygon triangulates given CCW polygon using ear clipping algorithm (takes
// O(n^2) time). Produces array of two-coordinate, CCW triangles, put one after
// another. Returns empty array and error when triangulation did not complete
// properly.
func Polygon(points []Point) ([]float64, error) {
	n := len(points)

	if n < 3 {
		return nil, errors.New("cannot triangulate less than three points")
	}

	// Allocate memory for all needed elements and initialize them by hand.
	elements := make([]Element, n)
	elements[0].Prev, elements[0].Next = &elements[n-1], &elements[1]
	elements[0].Point = points[0]
	for i := 1; i < n-1; i++ {
		elements[i].Prev, elements[i].Next = &elements[i-1], &elements[i+1]
		elements[i].Point = points[i]
	}
	elements[n-1].Prev, elements[n-1].Next = &elements[n-2], &elements[0]
	elements[n-1].Point = points[n-1]

	ear := &elements[0]

	// Any triangulation of simple polygon has n-2 triangles. Triangle has 3
	// two-dimensional coordinates.
	i, t := 0, make([]float64, (n-2)*6)

	stop := ear
	var prev, next *Element

	for ear.Prev != ear.Next {
		prev = ear.Prev
		next = ear.Next

		if isEar(ear) {
			if polygonArea([]Point{prev.Point, ear.Point, next.Point}) > 0 {
				t[i+0], t[i+1] = prev.Point.X, prev.Point.Y
				t[i+2], t[i+3] = ear.Point.X, ear.Point.Y
				t[i+4], t[i+5] = next.Point.X, next.Point.Y
				i += 6
			}

			ear.Remove()
			ear = ear.Next
			stop = ear
			continue
		}

		ear = next

		if ear == stop {
			return []float64{}, errors.New("oops")
		}
	}

	// Return array slice of size consisting only of the elements actually took by
	// the triangulation (sometimes the number of triangles is lower than n-2 and
	// zeroes are filling the rest of the array).
	return t[0:i], nil
}
