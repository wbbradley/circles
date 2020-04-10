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
	x, y, radius float64
	color        color.RGBA
}

type CircleTree struct {
	papa   *Circle
	babies []*CircleTree
}

const (
	imgSize        = 10000
	minCircleSize  = imgSize / 200.0
	subCircleCount = 100
)

func drawCircle(gc *draw2dimg.GraphicContext, rootCircle *Circle) {
	gc.BeginPath()
	gc.SetFillColor(rootCircle.color)
	draw2dkit.Circle(gc, rootCircle.x, rootCircle.y, rootCircle.radius)
	gc.Fill()
}

func distance(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
}

func dist(a, b *Circle) float64 {
	return distance(a.x, a.y, b.x, b.y)
}

func calcMaxRadiusFrom(x, y float64, siblings []*CircleTree) float64 {
	radius := 10000000.0
	for _, sibling := range siblings {
		siblingCenterDist := distance(x, y, sibling.papa.x, sibling.papa.y)
		if siblingCenterDist < sibling.papa.radius {
			return 0.0
		}
		radius = min(radius, siblingCenterDist-sibling.papa.radius)
	}
	return radius
}

func addCircle(tree *CircleTree, depth int) {
	var child *CircleTree = nil

	for {
		circle := &Circle{
			x: rand.Float64()*tree.papa.radius*2 - tree.papa.radius + tree.papa.x,
			y: rand.Float64()*tree.papa.radius*2 - tree.papa.radius + tree.papa.y,
			color: color.RGBA{uint8(rand.Float64()*192 + 32),
				uint8(rand.Float64()*192 + 32),
				uint8(rand.Float64()*192 + 32),
				0xff},
		}
		distance := dist(tree.papa, circle)
		if distance >= tree.papa.radius-minCircleSize {
			continue
		}
		maxRadius := min(distance, calcMaxRadiusFrom(circle.x, circle.y, tree.babies))
		if maxRadius < minCircleSize {
			continue
		}
		circle.radius = min(tree.papa.radius-distance, maxRadius)
		child = &CircleTree{
			papa:   circle,
			babies: nil,
		}
		tree.babies = append(tree.babies, child)
		break
	}
	if depth > 0 && child.papa.radius > minCircleSize*2.1 {
		addCircle(child, depth-1)
	}
}

func min(x, y float64) float64 {
	if x < y {
		return x
	} else {
		return y
	}
}

func populateTree(tree *CircleTree) {
	addCircle(tree, 10)

	for i := 0; i < subCircleCount; i++ {
		addCircle(tree, subCircleCount)
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
	pink := color.RGBA{0xff, 0xcc, 0xcc, 0xff}
	tree := &CircleTree{
		papa: &Circle{
			x:      width / 2.0,
			y:      height / 2.0,
			radius: width/2.0 - thickness*2.0,
			color:  pink,
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

	populateTree(tree)
	drawTree(gc, tree)

	// Save to file
	draw2dimg.SaveToPngFile("circle_tree.png", dest)

}
