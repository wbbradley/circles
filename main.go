package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
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
	imgSize        = 4000
	minCircleSize  = imgSize / 200.0
	subCircleCount = 100
)

func drawCircle(gc *draw2dimg.GraphicContext, rootCircle *Circle) {
	gc.BeginPath()
	// gc.SetFillColor(rootCircle.color)
	// gc.SetStrokeColor(rootCircle.color)
	gc.SetStrokeColor(color.RGBA{A: 0xff})
	draw2dkit.Circle(gc, rootCircle.center.x, rootCircle.center.y, rootCircle.radius)
	gc.Stroke()
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}

func dist(a, b *Circle) float64 {
	return distance(a.center.x, a.center.y, b.center.x, b.center.y)
}

func calcMaxRadiusFrom(x, y float64, siblings []*CircleTree) float64 {
	radius := 10000000.0
	for _, sibling := range siblings {
		siblingCenterDist := distance(x, y, sibling.papa.center.x, sibling.papa.center.y)
		if siblingCenterDist < sibling.papa.radius {
			return 0.0
		}
		radius = min(radius, siblingCenterDist-sibling.papa.radius)
	}
	return radius
}

func addCircle(tree *CircleTree, depth int, c1, c2 color.RGBA) {
	var child *CircleTree = nil

	for {
		circle := Circle{
			center: Vector2D{
				rand.Float64()*tree.papa.radius*2 - tree.papa.radius + tree.papa.center.x,
				rand.Float64()*tree.papa.radius*2 - tree.papa.radius + tree.papa.center.y,
			},
			color: c1,
		}

		distance := dist(tree.papa, &circle)
		if distance >= tree.papa.radius-minCircleSize {
			continue
		}
		maxRadius := min(distance, calcMaxRadiusFrom(circle.center.x,
			circle.center.y, tree.babies))
		if maxRadius < minCircleSize {
			continue
		}
		circle.radius = min(tree.papa.radius-distance, maxRadius)
		child = &CircleTree{
			papa:   &circle,
			babies: nil,
		}
		tree.babies = append(tree.babies, child)
		break
	}
	if depth > 0 && child.papa.radius > minCircleSize*2.1 {
		addCircle(child, depth-1, c2, c1)
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

func sub(a, b *Vector2D) Vector2D {
	return Vector2D{a.x - b.x, a.y - b.y}
}

func add(a, b *Vector2D) Vector2D {
	return Vector2D{a.x + b.x, a.y + b.y}
}

func normalize(a *Vector2D) {
	dist := math.Sqrt(a.x*a.x + a.y*a.y)
	a.x /= dist
	a.y /= dist
}

func radiusForFiller(p, outerRadius float64) float64 {
	return math.Sin(p) * outerRadius
}

func populateTree(tree *CircleTree, c1, c2 color.RGBA) {
	addCircle(tree, 0, c1, c2)
	papa := tree.papa
	child := tree.babies[0].papa
	vec := sub(&child.center, &papa.center)
	normalize(&vec)
	angle := math.Atan2(vec.y, vec.x)
	for theta := 0.01; theta < math.Pi*2; theta += math.Pi / 20.0 {
		p := angle + theta
		newCenter := add(&child.center, &Vector2D{
			math.Cos(p) * child.radius,
			math.Sin(p) * child.radius,
		})
		tree.babies = append(tree.babies, &CircleTree{
			papa: &Circle{
				center: newCenter,
				radius: radiusForFiller(theta, 100.0), //child.radius),
				color:  child.color,
			},
			babies: nil,
		})
	}
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
	var thickness float64 = 4.0
	width := float64(imgSize)
	height := float64(imgSize)
	white := color.RGBA{0xff, 0xff, 0xff, 0xff}
	black := color.RGBA{0x0, 0x0, 0x0, 0xff}
	tree := &CircleTree{
		papa: &Circle{
			center: Vector2D{
				x: width / 2.0,
				y: height / 2.0,
			},
			radius: width/2.0 - thickness*2.0,
			color:  black,
		},
		babies: nil,
	}

	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	gc := draw2dimg.NewGraphicContext(dest)

	// Clear the background
	gc.SetFillColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	gc.BeginPath()
	gc.MoveTo(0, 0)
	gc.LineTo(width, 0)
	gc.LineTo(width, height)
	gc.LineTo(0, height)
	gc.Close()
	gc.Fill()

	gc.SetLineWidth(thickness)

	populateTree(tree, white, black)
	drawTree(gc, tree)

	// Save to file
	draw2dimg.SaveToPngFile("circle_tree.png", dest)

}
