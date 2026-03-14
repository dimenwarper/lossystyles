package renderers

import "math"

// Vec3 is a 3D vector.
type Vec3 struct{ X, Y, Z float64 }

// Edge3D connects two vertex indices.
type Edge3D struct{ A, B int }

// Icosahedron returns the 12 vertices and 30 edges of a regular icosahedron.
func Icosahedron() ([]Vec3, []Edge3D) {
	phi := (1 + math.Sqrt(5)) / 2

	vertices := []Vec3{
		{0, 1, phi}, {0, -1, phi}, {0, 1, -phi}, {0, -1, -phi},
		{1, phi, 0}, {-1, phi, 0}, {1, -phi, 0}, {-1, -phi, 0},
		{phi, 0, 1}, {-phi, 0, 1}, {phi, 0, -1}, {-phi, 0, -1},
	}

	// Edge length² = 4.0 for this parameterization
	targetDistSq := 4.0
	eps := 0.1
	var edges []Edge3D
	for i := 0; i < len(vertices); i++ {
		for j := i + 1; j < len(vertices); j++ {
			dx := vertices[i].X - vertices[j].X
			dy := vertices[i].Y - vertices[j].Y
			dz := vertices[i].Z - vertices[j].Z
			distSq := dx*dx + dy*dy + dz*dz
			if math.Abs(distSq-targetDistSq) < eps {
				edges = append(edges, Edge3D{i, j})
			}
		}
	}

	return vertices, edges
}

// Octahedron returns the 6 vertices and 12 edges of a regular octahedron.
func Octahedron() ([]Vec3, []Edge3D) {
	vertices := []Vec3{
		{1, 0, 0}, {-1, 0, 0},
		{0, 1, 0}, {0, -1, 0},
		{0, 0, 1}, {0, 0, -1},
	}
	edges := []Edge3D{
		{0, 2}, {0, 3}, {0, 4}, {0, 5},
		{1, 2}, {1, 3}, {1, 4}, {1, 5},
		{2, 4}, {2, 5}, {3, 4}, {3, 5},
	}
	return vertices, edges
}

func rotateX(v Vec3, a float64) Vec3 {
	cos, sin := math.Cos(a), math.Sin(a)
	return Vec3{v.X, v.Y*cos - v.Z*sin, v.Y*sin + v.Z*cos}
}

func rotateY(v Vec3, a float64) Vec3 {
	cos, sin := math.Cos(a), math.Sin(a)
	return Vec3{v.X*cos + v.Z*sin, v.Y, -v.X*sin + v.Z*cos}
}

func rotateZ(v Vec3, a float64) Vec3 {
	cos, sin := math.Cos(a), math.Sin(a)
	return Vec3{v.X*cos - v.Y*sin, v.X*sin + v.Y*cos, v.Z}
}

// DrawPolytope renders a rotating icosahedron wireframe onto the canvas.
// The polytope slowly tumbles based on the step counter.
func DrawPolytope(canvas *Canvas, step int, edgeColor, vertexColor string) {
	vertices, edges := Icosahedron()

	// Slow tumble — different speeds per axis for organic feel
	t := float64(step) * 0.03
	ax := t * 0.7
	ay := t * 1.0
	az := t * 0.4

	centerX := float64(canvas.Width) / 2
	centerY := float64(canvas.Height) / 2

	// Scale to fit ~60% of the screen
	scale := math.Min(float64(canvas.Width)/5.5, float64(canvas.Height)/2.8)

	// Transform and project all vertices
	projected := make([][2]int, len(vertices))
	for i, v := range vertices {
		v = rotateX(v, ax)
		v = rotateY(v, ay)
		v = rotateZ(v, az)

		// Orthographic projection with terminal aspect ratio correction (~2:1)
		screenX := int(math.Round(centerX + v.X*scale))
		screenY := int(math.Round(centerY + v.Y*scale*0.5))
		projected[i] = [2]int{screenX, screenY}
	}

	// Draw edges as dotted lines — bolder dot character
	for _, e := range edges {
		a, b := projected[e.A], projected[e.B]
		canvas.DrawLine(a[0], a[1], b[0], b[1], '•', edgeColor)
	}

	// Draw vertices on top — bright and bold
	for _, p := range projected {
		canvas.SetBold(p[0], p[1], '◆', vertexColor)
	}
}

// DrawPolytopeOctahedron renders a rotating octahedron — simpler, bolder wireframe.
func DrawPolytopeOctahedron(canvas *Canvas, step int, edgeColor, vertexColor string) {
	vertices, edges := Octahedron()

	t := float64(step) * 0.025
	ax := t * 0.5
	ay := t * 0.8
	az := t * 0.3

	centerX := float64(canvas.Width) / 2
	centerY := float64(canvas.Height) / 2
	scale := math.Min(float64(canvas.Width)/4.5, float64(canvas.Height)/2.5)

	projected := make([][2]int, len(vertices))
	for i, v := range vertices {
		v = rotateX(v, ax)
		v = rotateY(v, ay)
		v = rotateZ(v, az)

		screenX := int(math.Round(centerX + v.X*scale))
		screenY := int(math.Round(centerY + v.Y*scale*0.5))
		projected[i] = [2]int{screenX, screenY}
	}

	for _, e := range edges {
		a, b := projected[e.A], projected[e.B]
		canvas.DrawLine(a[0], a[1], b[0], b[1], '·', edgeColor)
	}

	for _, p := range projected {
		canvas.SetBold(p[0], p[1], '◆', vertexColor)
	}
}
