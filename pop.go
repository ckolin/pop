package main

import (
	"flag"
	"image"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/fogleman/gg"
)

const populationSize = 1000
const geneLength = 6
const mutationRate = 0.01
const generations = 16384

type Shape struct {
	x, y, w, h float64
	r, g, b    float64
}

func generateShape() Shape {
	return Shape{
		x: rand.Float64(),
		y: rand.Float64(),
		w: rand.Float64(),
		h: rand.Float64(),
		r: rand.Float64(),
		g: rand.Float64(),
		b: rand.Float64(),
	}
}

func (s *Shape) Mutate() {
	if rand.Float64() < mutationRate {
		mutateFloat(&s.x)
	}
	if rand.Float64() < mutationRate {
		mutateFloat(&s.y)
	}
	if rand.Float64() < mutationRate {
		mutateFloat(&s.w)
	}
	if rand.Float64() < mutationRate {
		mutateFloat(&s.h)
	}
	if rand.Float64() < mutationRate {
		mutateFloat(&s.r)
	}
	if rand.Float64() < mutationRate {
		mutateFloat(&s.g)
	}
	if rand.Float64() < mutationRate {
		mutateFloat(&s.b)
	}
}

func mutateFloat(f *float64) {
	*f = math.Max(0, math.Min(1, *f+(rand.Float64()-0.5)))
}

type Dna struct {
	genes []Shape
	f     float64
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

func (a *Dna) Combine(b *Dna) Dna {
	c := Dna{
		genes: make([]Shape, len(a.genes)),
	}

	fac := rand.Float64()

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
		dna.genes[i].Mutate()
	}
}

func (dna *Dna) Render(dc *gg.Context) {
	bounds := dc.Image().Bounds()
	dc.Scale(float64(bounds.Dx()), float64(bounds.Dy()))
	for _, shape := range dna.genes {
		dc.DrawRectangle(
			shape.x-shape.w/2.0,
			shape.y-shape.h/2.0,
			shape.w,
			shape.h,
		)
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

	// Create blank image
	dc := gg.NewContext(bounds.Dx(), bounds.Dy())
	dc.SetRGB(0.5, 0.5, 0.5)
	dc.Clear()
	blank := dc.Image()

	rand.Seed(time.Now().UTC().UnixNano())

	// Initialize random population
	pop := make([]Dna, populationSize)
	for i := range pop {
		pop[i] = generateDna()
	}

	for gen := 1; gen <= generations; gen++ {
		// Calculate fitness
		var wg sync.WaitGroup
		wg.Add(len(pop))

		total := 0.0
		best := 0.0
		best_i := 0

		for i := range pop {
			go func(i int) {
				defer wg.Done()
				test := gg.NewContextForImage(blank)
				pop[i].Render(test)
				pop[i].f = fitness(test.Image(), goal)

				if pop[i].f > best {
					best = pop[i].f
					best_i = i
				}
				total += pop[i].f
			}(i)
		}

		wg.Wait()

		// Arrange for selection
		sum := 0.0
		for i := range pop {
			pop[i].f /= total
			sum += pop[i].f
			pop[i].f = sum
		}

		// Output
		if gen == generations || gen&(gen-1) == 0 {
			out := gg.NewContextForImage(dc.Image())
			pop[best_i].Render(out)
			out.SavePNG("out/gen" + strconv.Itoa(gen) + ".png")
			print(".")
		}

		// Create new population
		new := make([]Dna, len(pop))
		for i := range new {
			dna := pick(pop).Combine(pick(pop))
			dna.Mutate()
			new[i] = dna
		}
		pop = new
	}
}

func parseArgs() string {
	inPtr := flag.String("i", "in.jpg", "input file")
	flag.Parse()
	return *inPtr
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

func fitness(img, goal image.Image) float64 {
	bounds := goal.Bounds()
	count := bounds.Dx() * bounds.Dy()
	maxDist := dist(color.White, color.Black)
	score := 0.0
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			err := dist(img.At(x, y), goal.At(x, y)) / maxDist
			score += math.Pow(err, 2)
		}
	}
	score /= float64(count)
	return 1.0 / score
}

func dist(a, b color.Color) float64 {
	ra, ga, ba := floatParts(a)
	rb, gb, bb := floatParts(b)
	return math.Abs(rb-ra) + math.Abs(gb-ga) + math.Abs(bb-ba)
}

func floatParts(col color.Color) (float64, float64, float64) {
	m := 65536.0
	r, g, b, _ := col.RGBA()
	return float64(r) / m, float64(g) / m, float64(b) / m
}
