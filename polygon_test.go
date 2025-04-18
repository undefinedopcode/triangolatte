package triangolatte

import (
	"fmt"
	"math"
	"sort"
	"testing"
)

func TestPolygonArea(t *testing.T) {
	points := []Point{{2, 2}, {11, 2}, {9, 7}, {4, 10}}
	area := polygonArea(points)

	if area != 45.5 {
		t.Error("polygonArea implementation is wrong")
	}
}

func TestDeviation(t *testing.T) {
	data := []Point{{0, 4}, {3, 1}, {8, 2}, {9, 5}, {4, 6}}
	triangles := []float64{4, 6, 0, 4, 3, 1, 4, 6, 3, 1, 8, 2, 8, 2, 9, 5, 4, 6}

	actual, calculated, deviationResult := deviation(data, [][]Point{}, triangles)
	if deviationResult > 0 {
		t.Errorf("real: %f, actual: %f", actual, calculated)
	}
}

func checkFloat64Array(t *testing.T, result, expected []float64) {
	if len(result) != len(expected) {
		t.Error("Array sizes don't match")
	}

	for i, r := range result {
		if math.Abs(r-expected[i]) > 0.001 {
			t.Error("Value error beyond floating point precision problem")
		}
	}
}

func checkPointArray(t *testing.T, result, expected []Point) {
	if len(result) != len(expected) {
		t.Error("Array sizes don't match")
	}

	for i, r := range result {
		if math.Abs(r.X-expected[i].X) > 0.001 && math.Abs(result[i].Y-expected[i].Y) > 0.001 {
			t.Error("Value error beyond floating point precision problem")
		}
	}
}

func TestIsReflex(t *testing.T) {
	t.Run("convex", func(t *testing.T) {
		convex := []Point{{0, 1}, {1, 0}, {2, 1}}
		if isReflex(convex[0], convex[1], convex[2]) {
			t.Error("isReflex: false negative")
		}
	})

	t.Run("reflex", func(t *testing.T) {
		reflex := []Point{{0, 0}, {0, 3}, {2, 3}}
		if !isReflex(reflex[0], reflex[1], reflex[2]) {
			t.Error("isReflex: false positive")
		}
	})

	t.Run("square", func(t *testing.T) {
		square := []Point{{1, 1}, {0, 1}, {0, 0}}
		if isReflex(square[0], square[1], square[2]) {
			t.Error("isReflex: false negative")
		}
	})

	t.Run("another reflex", func(t *testing.T) {
		anotherReflex := []Point{{0, 0}, {2, 3}, {4, 2}}
		if !isReflex(anotherReflex[0], anotherReflex[1], anotherReflex[2]) {
			t.Error("isReflex: false positive")
		}
	})
}

func BenchmarkIsReflex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isReflex(Point{0, 0}, Point{1, 1}, Point{2, 0})
	}
}

func TestIsInsideTriangle(t *testing.T) {
	t.Run("outside", func(t *testing.T) {
		if isInsideTriangle(Point{0, 0}, Point{4, 0}, Point{4, 2}, Point{2, 2}) {
			t.Error("isInsideTriangle: point outside detected inside")
		}
	})

	t.Run("inside", func(t *testing.T) {
		if !isInsideTriangle(Point{0, 0}, Point{3, 0}, Point{3, 3}, Point{1, 1}) {
			t.Error("isInsideTriangle: point inside detected outside")
		}
	})

	t.Run("on edge", func(t *testing.T) {
		if !isInsideTriangle(Point{0, 2}, Point{6, 0}, Point{6, 2}, Point{2, 2}) {
			t.Error("isInsideTriangle: point on the edge reported as outside")
		}
	})
}

func BenchmarkIsInsideTriangle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isInsideTriangle(Point{50, 110}, Point{150, 30}, Point{240, 115}, Point{320, 65})
	}
}

func TestJoinHoles(t *testing.T) {
	type TestInfo struct {
		Name     string
		Points   [][]Point
		Expected []Point
	}

	testInfo := []TestInfo{{
		"square in square",
		[][]Point{
			{{0, 0}, {4, 0}, {4, 4}, {0, 4}},
			{{1, 1}, {1, 3}, {3, 3}, {3, 1}},
		},
		[]Point{{0, 0}, {4, 0}, {4, 4}, {3, 3}, {3, 1}, {1, 1}, {1, 3}, {3, 3}, {4, 4}, {0, 4}},
	}, {
		"triangle touching edge",
		[][]Point{
			{{0, 0}, {4, 0}, {4, 4}, {0, 4}},
			{{1, 1}, {1, 3}, {4, 2}},
		},
		[]Point{{0, 0}, {4, 2}, {1, 1}, {1, 3}, {4, 2}, {0, 0}, {4, 0}, {4, 4}, {0, 4}},
	}}

	for _, test := range testInfo {
		t.Run(test.Name, func(t *testing.T) {
			result, err := JoinHoles(test.Points)

			if err != nil {
				t.Errorf("JoinHoles: %s", err)
			}

			t.Log(result)

			checkPointArray(t, result, test.Expected)

			triangles, err := Polygon(result)

			if err != nil {
				t.Errorf("JoinHoles: %s", err)
			}

			t.Log(triangles)

			actual, calculated, deviationResult := deviation(test.Points[0], test.Points[1:], triangles)
			if deviationResult > 0 {
				t.Errorf("real: %f, actual: %f", actual, calculated)
			}
		})
	}

	t.Run("empty", func(t *testing.T) {
		_, err := JoinHoles([][]Point{})

		if err == nil {
			t.Error("JoinHoles: empty does not cause error")
		}
	})

	t.Run("only outer", func(t *testing.T) {
		points := []Point{{0.0, 0.0}, {1.0, 1.0}}
		result, err := JoinHoles([][]Point{points})

		if err != nil {
			t.Errorf("JoinHoles: %s", err)
		}

		checkPointArray(t, result, points)
	})
}

func TestPolygon(t *testing.T) {
	type TestInfo struct {
		Name  string
		Shape []Point
	}

	shapes := []TestInfo{
		{"fan", []Point{{0, 4}, {3, 1}, {8, 2}, {9, 5}, {4, 6}}},
		{"diamond", []Point{{0, 3}, {1, 0}, {4, 1}, {3, 4}}},
		{"square", []Point{{0, 0}, {1, 0}, {1, 1}, {0, 1}}},
		{"one reflex", []Point{{0, 6}, {0, 1}, {2, 2}, {3, 2}}},
		{"shuriken", []Point{{0, 4}, {2, 2}, {2, 0}, {4, 2}, {6, 2}, {4, 4}, {4, 6}, {2, 4}}},
		{"c letter", []Point{{0, 0}, {4, 0}, {4, 2}, {2, 2}, {2, 4}, {4, 4}, {4, 6}, {0, 6}}},
		{"t letter", []Point{{0, 0}, {6, 0}, {6, 2}, {4, 2}, {4, 6}, {2, 6}, {2, 2}, {0, 2}}},
		{"double t", []Point{{0, 0}, {6, 0}, {6, 2}, {4, 2}, {4, 4}, {6, 4}, {6, 6}, {0, 6}, {0, 4}, {2, 4}, {2, 2}, {0, 2}}},
		{"building", []Point{{1, 0}, {7, 0}, {7, 1}, {6, 1}, {6, 10}, {7, 10}, {7, 11}, {1, 11}, {1, 10}, {2, 10}, {2, 7}, {0, 7}, {0, 4}, {2, 4}, {2, 1}, {1, 1}}},
		{"from the paper", []Point{{50, 110}, {150, 30}, {240, 115}, {320, 65}, {395, 170}, {305, 160}, {265, 240}, {190, 100}, {95, 125}, {100, 215}}},
	}

	for _, s := range shapes {
		t.Run(fmt.Sprintf("%s", s.Name), func(t *testing.T) {
			res, err := Polygon(s.Shape)
			if err != nil {
				t.Error(err)
			}

			actual, calculated, dif := deviation(s.Shape, [][]Point{}, res)

			if dif != 0 {
				t.Errorf("#%s: real area: %f; result: %f", s.Name, actual, calculated)
			}
			t.Logf("#%s: %v", s.Name, res)
		})
	}
}

func BenchmarkPolygon(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Polygon([]Point{{50, 110}, {150, 30}, {240, 115}, {320, 65}, {395, 170}, {305, 160}, {265, 240}, {190, 100}, {95, 125}, {100, 215}})
	}
}

func TestIncorrectPolygon(t *testing.T) {
	var err error
	_, err = Polygon([]Point{{0, 0}})
	if err == nil {
		t.Errorf("The code did not return error on incorrect input")
	}
}

func TestSortingByXMax(t *testing.T) {
	inners := [][]Point{{{1, 2}}, {{0, 0}}}
	sort.Sort(byMaxX(inners))
}

func TestSingleTriangleTriangulation(t *testing.T) {
	result, _ := Polygon([]Point{{0, 0}, {3, 0}, {4, 4}})
	expected := []float64{4, 4, 0, 0, 3, 0}

	t.Log(result)
	t.Log(expected)
	checkFloat64Array(t, result, expected)
}

func TestAghA0(t *testing.T) {
	agh, _ := loadPointsFromFile("assets/agh_a0")
	for i := range agh {
		for j := range agh[i] {
			p := degreesToMeters(agh[i][j])
			agh[i][j] = Point{3 * (p.X - 2217750), 3 * (p.Y - 6457350)}
		}
	}

	result, err := Polygon(agh[0]) // agh[1:]

	if err != nil {
		t.Errorf("AghA0: %s", err)
	}

	actual, calculated, deviationResult := deviation(agh[0], [][]Point{}, result)
	if deviationResult > 1e-10 {
		t.Errorf("real area: %f; result: %f", actual, calculated)
	}
}

// Runs much longer than others (around half a minute)
// func TestLakeSuperior(t *testing.T) {
// 	lakeSuperior, _ := LoadPointsFromFile("assets/lake_superior")
//
// 	for i := range lakeSuperior {
// 		for j := range lakeSuperior[i] {
// 			p := degreesToMeters(lakeSuperior[i][j])
// 			lakeSuperior[i][j] = Point{math.Abs(p.X), math.Abs(p.Y)}
// 		}
// 	}
//
// 	result, err := Polygon(lakeSuperior[0]) // lakeSuperior[1:]
//
// 	if err != nil {
// 		t.Errorf("LakeSuperior: %s", err)
// 	}
//
// 	t.Log(result)
// }
//
// func TestFromFile(t *testing.T) {
// 	points, err := LoadPointsFromFile("assets/lake")
//
// 	if err != nil {
// 		t.Errorf("FromFile: %s", err)
// 	}
//
// 	for i := range points {
// 		for j := range points[i] {
// 			p := degreesToMeters(points[i][j])
// 			points[i][j] = p
// 		}
// 	}
//
// 	res, err := Polygon(points[0])
//
// 	if err != nil {
// 		t.Errorf("FromFile: %s", err)
// 	}
//
// 	actual, calculated, deviationResult := deviation(points[0], [][]Point{}, res)
// 	if deviationResult > 1e-10 {
// 		t.Errorf("real area: %f; result: %f", actual, calculated)
// 	}
//
// 	t.Log(res)
// }
