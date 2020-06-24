package main

import (
	"flag"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/fogleman/gg"
)

const populationSize = 400
const geneLength = 16
const mutationRate = 0.02
const generations = 1000

type Shape struct {
	x, y, s float64
	r, g, b float64
}

type Dna struct {
	genes []Shape
	f     float64
}

func (a *Dna) Combine(b *Dna) Dna {
	fac := rand.Float64()

	c := Dna{
		genes: make([]Shape, len(a.genes)),
	}

	for i := 0; i < geneLength; i++ {
		if rand.Float64() < fac {
			c.genes[i] = a.genes[i]
		} else {
			c.genes[i] = b.genes[i]
		}
	}

	return c
}

func (dna *Dna) Mutate() {
	for i := range dna.genes {
		if rand.Float64() < mutationRate {
			dna.genes[i] = generateShape()
		}
	}
}

func (dna *Dna) Render(dc *gg.Context) {
	bounds := dc.Image().Bounds()
	for i := range dna.genes {
		shape := dna.genes[i]
		dc.DrawCircle(shape.x*float64(bounds.Dx()), shape.y*float64(bounds.Dy()), shape.s*float64(bounds.Dx()))
		dc.SetRGBA(shape.r, shape.g, shape.b, 0.6)
		dc.Fill()
	}
}

func main() {
	in := parseArgs()

	goal, err := gg.LoadJPG(in)
	if err != nil {
		panic(err)
	}
	bounds := goal.Bounds()

	// Create starting canvas
	dc := gg.NewContext(bounds.Dx(), bounds.Dy())
	dc.SetRGB(0, 0, 0)
	dc.Clear()

	rand.Seed(time.Now().UTC().UnixNano())

	// Initialize random population
	pop := make([]Dna, populationSize)
	for i := range pop {
		pop[i] = generateDna()
	}

	for gen := 1; gen <= generations; gen++ {
		// Calculate fitness
		total := 0.0
		for i := range pop {
			test := gg.NewContextForImage(dc.Image())
			pop[i].Render(test)
			pop[i].f = rate(test.Image(), goal)
			total += pop[i].f
		}
		sum := 0.0
		for i := range pop {
			pop[i].f /= total
			sum += pop[i].f
			pop[i].f = sum
		}
		sort.Slice(pop, func(i, j int) bool {
			return pop[i].f < pop[j].f
		})

		if gen%10 == 0 {
			out := gg.NewContextForImage(dc.Image())
			pop[populationSize-1].Render(out)
			out.SavePNG("out/gen" + strconv.Itoa(gen) + ".png")
		}

		// Create new population
		new := make([]Dna, len(pop))
		for i := range new {
			// Pick two parents
			mom := pick(pop)
			dad := pick(pop)
			dna := mom.Combine(dad)
			dna.Mutate()
			new[i] = dna
		}
		pop = new
	}
}

func generateDna() Dna {
	dna := Dna{
		genes: make([]Shape, geneLength),
	}

	for i := range dna.genes {
		dna.genes[i] = generateShape()
	}

	return dna
}

func generateShape() Shape {
	return Shape{
		x: rand.Float64(),
		y: rand.Float64(),
		s: rand.Float64() * 0.5,
		r: rand.Float64(),
		g: rand.Float64(),
		b: rand.Float64(),
	}
}

func pick(pop []Dna) *Dna {
	r := rand.Float64()
	for _, dna := range pop {
		if dna.f >= r {
			return &dna
		}
	}
	panic("error in fitness calculation")
}

func rate(img, goal image.Image) float64 {
	bounds := goal.Bounds()
	count := bounds.Dx() * bounds.Dy()
	maxDist := dist(color.White, color.Black)
	score := 0.0
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			err := dist(img.At(x, y), goal.At(x, y)) / maxDist
			score += 1.0 - err
		}
	}
	score /= float64(count)
	return math.Pow(score, 2)
}

func dist(a, b color.Color) float64 {
	ra, ga, ba := floatParts(a)
	rb, gb, bb := floatParts(b)
	return math.Pow(float64(rb-ra), 2) + math.Pow(float64(gb-ga), 2) + math.Pow(float64(bb-ba), 2)
}

func floatParts(col color.Color) (float64, float64, float64) {
	r, g, b, _ := col.RGBA()
	return float64(r) / 65536.0, float64(g) / 65536.0, float64(b) / 65536.0
}

func parseArgs() string {
	inPtr := flag.String("i", "in.jpg", "input file")
	flag.Parse()
	return *inPtr
}
