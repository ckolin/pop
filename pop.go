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

type Dna struct {
	x, y, s float64
	r, g, b float64
	f       float64
}

func mid(a, b float64) float64 {
	return (a + b) / 2.0
}

func (a *Dna) Combine(b *Dna) Dna {
	return Dna{
		x: mid(a.x, b.x),
		y: mid(a.y, b.y),
		s: mid(a.s, b.s),
		r: mid(a.r, b.r),
		g: mid(a.g, b.g),
		b: mid(a.b, b.b),
	}
}

func mut(val float64) float64 {
	val += (rand.Float64() - 0.5) * 0.2
	return math.Min(1.0, math.Max(0.0, val))
}

func (dna *Dna) Mutate() {
	// Change values randomly
	dna.x = mut(dna.x)
	dna.y = mut(dna.y)
	dna.s = mut(dna.s)
	dna.r = mut(dna.r)
	dna.g = mut(dna.g)
	dna.b = mut(dna.b)
}

func (dna *Dna) Render(dc *gg.Context) {
	bounds := dc.Image().Bounds()
	dc.DrawCircle(dna.x*float64(bounds.Dx()), dna.y*float64(bounds.Dy()), dna.s*float64(bounds.Dx()))
	dc.SetRGBA(dna.r, dna.g, dna.b, 0.9)
	dc.Fill()
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

	for element := 0; element < 30; element++ {
		// Initialize random population
		pop := make([]Dna, 20)
		for i := 0; i < len(pop); i++ {
			pop[i] = Dna{
				x: rand.Float64(),
				y: rand.Float64(),
				s: rand.Float64(),
				r: rand.Float64(),
				g: rand.Float64(),
				b: rand.Float64(),
			}
		}

		for gen := 0; gen < 100; gen++ {
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

		choice := pop[rand.Int()%len(pop)]
		choice.Render(dc)
		dc.SavePNG("out/" + strconv.Itoa(element) + ".png")
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
			error := dist(img.At(x, y), goal.At(x, y)) / maxDist
			score += 1.0 - math.Pow(error, 2)
		}
	}
	return score / float64(count)
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
