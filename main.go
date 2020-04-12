package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"github.com/lucasb-eyer/go-colorful"
)

type Circle struct {
	center Vector2D
	radius float64
	color  color.RGBA
}

type CircleTree struct {
	papa   *Circle
	babies []*CircleTree
}

const (
	imgSize        = 2000
	maxDepth       = 12
	maxRadiusRatio = 0.95
	minRadiusRatio = 0.5
	increment      = 0.125 / 2.0
)

var (
	treeNum       = 0
	minCircleSize = 4.0
	thickness     = 1.0
	palette       = genPalette(maxDepth) // colorful.WarmPalette(maxDepth)
)

func drawCircle(gc *draw2dimg.GraphicContext, c *Circle) {
	gc.BeginPath()
	gc.SetFillColor(c.color)
	// gc.SetStrokeColor(color.RGBA{0, 0, 0, 255}) //c.color)
	draw2dkit.Circle(gc, c.center.x, c.center.y, c.radius)
	gc.Fill()
}

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

func genPalette(d int) []colorful.Color {
	keypoints := GradientTable{
		{MustParseHex("#9e0142"), 0.0},
		{MustParseHex("#d53e4f"), 0.1},
		{MustParseHex("#f46d43"), 0.2},
		{MustParseHex("#fdae61"), 0.3},
		{MustParseHex("#fee090"), 0.4},
		{MustParseHex("#ffffbf"), 0.5},
		{MustParseHex("#e6f598"), 0.6},
		{MustParseHex("#abdda4"), 0.7},
		{MustParseHex("#66c2a5"), 0.8},
		{MustParseHex("#3288bd"), 0.9},
		{MustParseHex("#5e4fa2"), 1.0},
	}
	p := make([]colorful.Color, 0, d)
	for i := 0; i < d; i++ {
		p = append(p, keypoints.GetInterpolatedColorFor(float64(i)/float64(d)))
	}
	return p
}

func MustParseHex(s string) colorful.Color {
	c, err := colorful.Hex(s)
	if err != nil {
		panic("MustParseHex: " + err.Error())
	}
	return c
}

func (self GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(self)-1; i++ {
		c1 := self[i]
		c2 := self[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return self[len(self)-1].Col
}

func pointDistance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}

func distance(a, b Vector2D) float64 {
	return pointDistance(a.x, a.y, b.x, b.y)
}

func dist(a, b *Circle) float64 {
	return pointDistance(a.center.x, a.center.y, b.center.x, b.center.y)
}

func calcMaxRadiusFrom(x, y float64, siblings []*CircleTree) float64 {
	radius := 10000000.0
	for _, sibling := range siblings {
		siblingCenterDist := pointDistance(x, y, sibling.papa.center.x, sibling.papa.center.y)
		if siblingCenterDist < sibling.papa.radius {
			return 0.0
		}
		radius = min(radius, siblingCenterDist-sibling.papa.radius)
	}
	return radius
}

func sqr(x float64) float64 {
	return x * x
}

func HSVtoRGBA(x, y, z float64) color.RGBA {
	c := colorful.Hsv(x, y, z)
	return color.RGBA{
		R: uint8(math.Floor(c.R * 255.0)),
		G: uint8(math.Floor(c.G * 255.0)),
		B: uint8(math.Floor(c.B * 255.0)),
		A: 0xff,
	}
}

func ToRGBA(c colorful.Color) color.RGBA {
	return color.RGBA{
		uint8(c.R * 255.0),
		uint8(c.G * 255.0),
		uint8(c.B * 255.0),
		0xff,
	}
}

func randColor(depth int) color.RGBA {
	if depth >= len(palette) {
		panic("invalid depth")
	}
	return ToRGBA(palette[depth])
}

func circlesIntersect(a, b *Circle) (Vector2D, Vector2D) {
	r1_squared := sqr(a.radius)
	r2_squared := sqr(b.radius)
	R := distance(a.center, b.center)
	R_squared := sqr(R)
	base := add(
		midpoint(a.center, b.center),
		scale(
			(r1_squared-r2_squared)/(2.0*R_squared),
			sub(b.center, a.center)))
	C := 0.5 * math.Sqrt(2.0*(r1_squared+r2_squared)/R_squared-sqr(r1_squared-r2_squared)/sqr(R_squared)-1.0)
	offset := scale(C, Vector2D{b.center.y - a.center.y, a.center.x - b.center.x})
	return add(base, offset), sub(base, offset)
}

func (t *CircleTree) validCircle(c *Circle) bool {
	if c.radius < minCircleSize {
		return false
	}
	if distance(t.papa.center, c.center) > t.papa.radius-c.radius+2.0 {
		return false
	}
	for _, child := range t.babies {
		if distance(child.papa.center, c.center) < c.radius+child.papa.radius-2.0 {
			return false
		}
	}
	return true
}

func lerp(a, x0, x1 float64) float64 {
	return (x1-x0)*a + x0
}

func addCircle(tree *CircleTree, depth int) bool {
	var child *CircleTree = nil

	for i := 0; i < 1000; i += 1 {
		theta := rand.Float64() * math.Pi * 2.0
		radius := lerp(rand.Float64(), minRadiusRatio, maxRadiusRatio) * tree.papa.radius
		circle := Circle{
			center: Vector2D{
				math.Cos(theta)*(tree.papa.radius-radius) + tree.papa.center.x,
				math.Sin(theta)*(tree.papa.radius-radius) + tree.papa.center.y,
			},
			radius: radius,
			color:  randColor(depth),
		}

		if tree.validCircle(&circle) {
			child = &CircleTree{
				papa:   &circle,
				babies: nil,
			}
			tree.babies = append(tree.babies, child)
			return true
		}
	}
	return false
}

func max(x, y float64) float64 {
	if x > y {
		return x
	} else {
		return y
	}
}
func min(x, y float64) float64 {
	if x < y {
		return x
	} else {
		return y
	}
}

type Vector2D struct {
	x, y float64
}

func sub(a, b Vector2D) Vector2D {
	return Vector2D{a.x - b.x, a.y - b.y}
}

func add(a, b Vector2D) Vector2D {
	return Vector2D{a.x + b.x, a.y + b.y}
}

func normalize(a *Vector2D) {
	dist := math.Sqrt(a.x*a.x + a.y*a.y)
	a.x /= dist
	a.y /= dist
}

func norm(a Vector2D) Vector2D {
	dist := math.Sqrt(a.x*a.x + a.y*a.y)
	return Vector2D{a.x / dist, a.y / dist}
}

func radiusForFiller(p, outerRadius float64) float64 {
	return math.Sin(p) * outerRadius
}

func invert(p Vector2D) Vector2D {
	return Vector2D{-p.x, -p.y}
}

func scale(s float64, p Vector2D) Vector2D {
	return Vector2D{s * p.x, s * p.y}
}

func midpoint(a, b Vector2D) Vector2D {
	return scale(0.5, add(a, b))
}

func populateTree(tree *CircleTree, depth int) bool {
	if depth >= maxDepth {
		return false
	}
	if !addCircle(tree, depth) {
		return false
	}
	treeNum += 1
	fmt.Printf("Populating tree %d...\r", treeNum)
	papa := tree.papa
	childTree := tree.babies[len(tree.babies)-1]
	child := childTree.papa
	populateTree(childTree, depth+1)

	// Compute radius of filling circle
	b := papa.radius - child.radius
	// fmt.Printf("papa.radius = %v\nchild.radius = %v\nb = %v\n", papa.radius, child.radius, b)
	for r := 1.0; r < b; r += increment {
		shrunkPapa := Circle{
			center: papa.center,
			radius: papa.radius - r,
		}
		extendedChild := Circle{
			center: child.center,
			radius: child.radius + r,
		}
		p1, _ := circlesIntersect(&shrunkPapa, &extendedChild)

		c1 := &Circle{
			center: p1,
			radius: r,
		}

		if tree.validCircle(c1) {
			c1.color = randColor(depth)
			newTree := &CircleTree{
				papa:   c1,
				babies: nil,
			}
			tree.babies = append(tree.babies, newTree)
			populateTree(newTree, depth+1)
		}
	}
	for r := b - 1.0; r >= 1.0; r -= increment {
		shrunkPapa := Circle{
			center: papa.center,
			radius: papa.radius - r,
		}
		extendedChild := Circle{
			center: child.center,
			radius: child.radius + r,
		}
		_, p2 := circlesIntersect(&shrunkPapa, &extendedChild)
		c2 := &Circle{
			center: p2,
			radius: r,
		}

		if tree.validCircle(c2) {
			c2.color = randColor(depth)
			newTree := &CircleTree{
				papa:   c2,
				babies: nil,
			}
			tree.babies = append(tree.babies, newTree)
			populateTree(newTree, depth+1)
		}
	}
	return true
}

func drawTree(gc *draw2dimg.GraphicContext, tree *CircleTree) {
	if tree == nil {
		return
	}
	drawCircle(gc, tree.papa)
	for _, baby := range tree.babies {
		drawTree(gc, baby)
	}
}

func main() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
	width := float64(imgSize)
	height := float64(imgSize)
	white := color.RGBA{0xff, 0xff, 0xff, 0xff}
	// black := color.RGBA{0x0, 0x0, 0x0, 0xff}
	tree := &CircleTree{
		papa: &Circle{
			center: Vector2D{
				x: width / 2.0,
				y: height / 2.0,
			},
			radius: width/2.0 - thickness*2.0,
			color:  randColor(0),
		},
		babies: nil,
	}

	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Clear the background
	gc.SetFillColor(white)
	gc.BeginPath()
	gc.MoveTo(0, 0)
	gc.LineTo(width, 0)
	gc.LineTo(width, height)
	gc.LineTo(0, height)
	gc.Close()
	gc.Fill()

	gc.SetLineWidth(thickness)

	for {
		if !populateTree(tree, 1) {
			break
		}
	}
	fmt.Printf("\n")
	drawTree(gc, tree)

	// Save to file
	draw2dimg.SaveToPngFile("circle_tree.png", dest)

}
