// Copyright 2016 The Gosl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package msh

import (
	"math"

	"github.com/cpmech/gosl/chk"
	"github.com/cpmech/gosl/fun"
	"github.com/cpmech/gosl/num"
	"github.com/cpmech/gosl/plt"
	"github.com/cpmech/gosl/utl"
)

// IntPoint holds data of one integration (quadrature) point
type IntPoint struct {
	X []float64 // coordinates [ndim]
	W float64   // weight
}

// IntPoints implements integration points generate according to some rule; e.g. Gauss-Legendre
type IntPoints struct {
	Rule   string      // the rule; e.g. LE, LO, W5
	Ndim   int         // space dimension
	Npts   int         // number of points
	Points []*IntPoint // quadrature points
}

// rules:
//    LE -- Gauss-Legendre
//    LO -- Gauss-Lobatto
//    W5 -- Gauss-Legendre, Wilson's method with 5 points and variable weight
func NewIntPoints(rule string, ndim, npts int, prms fun.Params) (o *IntPoints) {

	o = new(IntPoints)
	o.Rule = rule
	o.Ndim = ndim
	o.Npts = npts
	o.Points = make([]*IntPoint, npts)

	switch rule {

	case "LE":
		n1d := int(math.Floor(math.Pow(float64(npts), 1.0/float64(ndim)) + 0.5))
		x, w := num.GaussLegendreXW(-1, 1, n1d)
		switch ndim {
		case 1:
			for i := 0; i < npts; i++ {
				o.Points[i] = &IntPoint{X: []float64{x[i]}, W: w[i]}
			}
		case 2:
			for j := 0; j < n1d; j++ {
				for i := 0; i < n1d; i++ {
					m := i + n1d*j
					o.Points[m] = &IntPoint{X: []float64{x[i], x[j]}, W: w[i] * w[j]}
				}
			}
		case 3:
			for k := 0; k < n1d; k++ {
				for j := 0; j < n1d; j++ {
					for i := 0; i < n1d; i++ {
						m := i + n1d*j + (n1d*n1d)*k
						o.Points[m] = &IntPoint{X: []float64{x[i], x[j], x[k]}, W: w[i] * w[j] * w[k]}
					}
				}
			}
		}

	case "W5corner", "W4stable", "W5":
		if ndim != 2 || npts != 5 {
			chk.Panic("rule %q works only with ndim=2 and npts=5. ndim=%d or npts=%d is invalid", ndim, npts)
		}
		w0 := 8.0 / 3.0
		wa := 1.0 / 3.0
		a := 1.0
		if rule == "W4stable" {
			w0 = 0.004
			wa = 0.999
			a = 0.5776391
		}
		if rule == "W5" {
			w0prm := prms.Find("w0")
			if w0prm == nil {
				chk.Panic("rule %q requires parameter w0 in prms", rule)
			}
			w0 = w0prm.V
			wa = (4.0 - w0) / 4.0
			a = math.Sqrt(1.0 / (3.0 * wa))
		}
		o.Points = []*IntPoint{
			{X: []float64{-a, -a}, W: wa},
			{X: []float64{+a, -a}, W: wa},
			{X: []float64{+0, +0}, W: w0},
			{X: []float64{-a, +a}, W: wa},
			{X: []float64{+a, +a}, W: wa},
		}

	case "W8fixed", "W8": // Appendix G-6, Eqs. (G.20)
		if ndim != 2 || npts != 8 {
			chk.Panic("rule %q works only with ndim=2 and npts=8. ndim=%d or npts=%d is invalid", ndim, npts)
		}
		a := math.Sqrt(7.0 / 9.0)
		b := math.Sqrt(7.0 / 15.0)
		wa := 9.0 / 49.0
		wb := 40.0 / 49.0
		if rule == "W8" {
			wbPrm := prms.Find("wb")
			if wbPrm == nil {
				chk.Panic("rule %q requires parameter wb in prms", rule)
			}
			wb = wbPrm.V
			wa = 1.0 - wb
			swa := math.Sqrt(wa)
			a = 1.0 / math.Sqrt(3.0*swa)
			b = math.Sqrt((2.0 - 2.0*swa) / (3.0 * wb))
		}
		o.Points = []*IntPoint{
			{X: []float64{-a, -a}, W: wa},
			{X: []float64{+0, -b}, W: wb},
			{X: []float64{+a, -a}, W: wa},
			{X: []float64{-b, +0}, W: wb},
			{X: []float64{+b, +0}, W: wb},
			{X: []float64{-a, +a}, W: wa},
			{X: []float64{+0, +b}, W: wb},
			{X: []float64{+a, +a}, W: wa},
		}

	default:
		chk.Panic("rule %q is not available", rule)
	}
	return
}

func (o IntPoints) Draw(dx []float64, args *plt.A) {
	if args == nil {
		args = &plt.A{C: "r", M: "*", Mec: "r", NoClip: true}
	}
	if dx == nil {
		dx = []float64{0, 0}
	}
	if o.Ndim == 2 {
		plt.Polyline([][]float64{
			{dx[0] - 1, dx[1] - 1}, {dx[0] + 1, dx[1] - 1}, {dx[0] + 1, dx[1] + 1}, {dx[0] - 1, dx[1] + 1},
		}, &plt.A{Fc: "none", Ec: "#2645cb", Closed: true, NoClip: true})
		for _, pts := range o.Points {
			plt.PlotOne(dx[0]+pts.X[0], dx[1]+pts.X[1], args)
		}
	}
}

// sets of integration points /////////////////////////////////////////////////////////////////////

// IntPointsSet implements a set of IntPoints (integration points); e.g. for "lin"
type IntPointsSet []*IntPoints

// Find finds sets of integration points
func (o *IntPointsSet) Find(rule string, npts int) (pts *IntPoints) {
	for _, pts = range *o {
		if pts.Rule == rule && pts.Npts == npts {
			return
		}
	}
	return nil
}

// linIntPoints holds all integration points currently generated for "lin" elements
var linIntPointsSet IntPointsSet

// quaIntPoints holds all integration points currently generated for "qua" elements
var quaIntPointsSet IntPointsSet

// hexIntPoints holds all integration points currently generated for "hex" elements
var hexIntPointsSet IntPointsSet

// initialise variables
func init() {

	// lin
	for n := 1; n <= 5; n++ {
		linIntPointsSet = append(linIntPointsSet, NewIntPoints("LE", 1, n, nil))
	}

	// qua
	quaIntPointsSet = append(quaIntPointsSet, NewIntPoints("LE", 2, 4, nil))
	quaIntPointsSet = append(quaIntPointsSet, NewIntPoints("LE", 2, 9, nil))
	quaIntPointsSet = append(quaIntPointsSet, NewIntPoints("W5corner", 2, 5, nil))
	quaIntPointsSet = append(quaIntPointsSet, NewIntPoints("W4stable", 2, 5, nil))
	quaIntPointsSet = append(quaIntPointsSet, NewIntPoints("W8fixed", 2, 8, nil))

	// hex
	hexIntPointsSet = append(hexIntPointsSet, NewIntPoints("LE", 3, 8, nil))
	hexIntPointsSet = append(hexIntPointsSet, NewIntPoints("LE", 3, 27, nil))
}

/////////////////////////////////////////////////////////////////////////////////////

// QuadPoint implements a quadrature (e.g. Gauss) point
type QuadPoint []float64 // length=4: [r, s, t, weight]

// QuadPoints is a set of quadrature points
type QuadPoints []QuadPoint

var IntPointsOld map[string]map[int]QuadPoints

// constants
func init() {

	SQ19by30 := math.Sqrt(19.0 / 30.0)
	SQ19by33 := math.Sqrt(19.0 / 33.0)

	IntPointsOld = make(map[string]map[int]QuadPoints)

	IntPointsOld["lin"] = map[int]QuadPoints{
		1: []QuadPoint{
			QuadPoint{0, 0, 0, 2},
		},
		2: []QuadPoint{
			QuadPoint{-0.5773502691896257, 0, 0, 1},
			QuadPoint{+0.5773502691896257, 0, 0, 1},
		},
		3: []QuadPoint{
			QuadPoint{-0.7745966692414834, 0, 0, 0.5555555555555556},
			QuadPoint{+0.0000000000000000, 0, 0, 0.8888888888888888},
			QuadPoint{+0.7745966692414834, 0, 0, 0.5555555555555556},
		},
		4: []QuadPoint{
			QuadPoint{-0.8611363115940526, 0, 0, 0.3478548451374538},
			QuadPoint{-0.3399810435848562, 0, 0, 0.6521451548625462},
			QuadPoint{+0.3399810435848562, 0, 0, 0.6521451548625462},
			QuadPoint{+0.8611363115940526, 0, 0, 0.3478548451374538},
		},
		5: []QuadPoint{
			QuadPoint{-0.9061798459386640, 0, 0, 0.2369268850561891},
			QuadPoint{-0.5384693101056831, 0, 0, 0.4786286704993665},
			QuadPoint{+0.0000000000000000, 0, 0, 0.5688888888888889},
			QuadPoint{+0.5384693101056831, 0, 0, 0.4786286704993665},
			QuadPoint{+0.9061798459386640, 0, 0, 0.2369268850561891},
		},
	}

	IntPointsOld["qua"] = map[int]QuadPoints{
		4: []QuadPoint{
			QuadPoint{-0.5773502691896257, -0.5773502691896257, 0, 1},
			QuadPoint{+0.5773502691896257, -0.5773502691896257, 0, 1},
			QuadPoint{-0.5773502691896257, +0.5773502691896257, 0, 1},
			QuadPoint{+0.5773502691896257, +0.5773502691896257, 0, 1},
		},
		9: []QuadPoint{
			QuadPoint{-0.7745966692414834, -0.7745966692414834, 0, 25.0 / 81.0},
			QuadPoint{+0.0000000000000000, -0.7745966692414834, 0, 40.0 / 81.0},
			QuadPoint{+0.7745966692414834, -0.7745966692414834, 0, 25.0 / 81.0},
			QuadPoint{-0.7745966692414834, +0.0000000000000000, 0, 40.0 / 81.0},
			QuadPoint{+0.0000000000000000, +0.0000000000000000, 0, 64.0 / 81.0},
			QuadPoint{+0.7745966692414834, +0.0000000000000000, 0, 40.0 / 81.0},
			QuadPoint{-0.7745966692414834, +0.7745966692414834, 0, 25.0 / 81.0},
			QuadPoint{+0.0000000000000000, +0.7745966692414834, 0, 40.0 / 81.0},
			QuadPoint{+0.7745966692414834, +0.7745966692414834, 0, 25.0 / 81.0},
		},
	}

	IntPointsOld["hex"] = map[int]QuadPoints{
		8: []QuadPoint{
			QuadPoint{-0.5773502691896257, -0.5773502691896257, -0.5773502691896257, 1},
			QuadPoint{+0.5773502691896257, -0.5773502691896257, -0.5773502691896257, 1},
			QuadPoint{-0.5773502691896257, +0.5773502691896257, -0.5773502691896257, 1},
			QuadPoint{+0.5773502691896257, +0.5773502691896257, -0.5773502691896257, 1},
			QuadPoint{-0.5773502691896257, -0.5773502691896257, +0.5773502691896257, 1},
			QuadPoint{+0.5773502691896257, -0.5773502691896257, +0.5773502691896257, 1},
			QuadPoint{-0.5773502691896257, +0.5773502691896257, +0.5773502691896257, 1},
			QuadPoint{+0.5773502691896257, +0.5773502691896257, +0.5773502691896257, 1},
		},
		14: []QuadPoint{
			QuadPoint{SQ19by30, 0.0, 0.0, 320.0 / 361.0},
			QuadPoint{-SQ19by30, 0.0, 0.0, 320.0 / 361.0},
			QuadPoint{0.0, SQ19by30, 0.0, 320.0 / 361.0},
			QuadPoint{0.0, -SQ19by30, 0.0, 320.0 / 361.0},
			QuadPoint{0.0, 0.0, SQ19by30, 320.0 / 361.0},
			QuadPoint{0.0, 0.0, -SQ19by30, 320.0 / 361.0},
			QuadPoint{SQ19by33, SQ19by33, SQ19by33, 121.0 / 361.0},
			QuadPoint{-SQ19by33, SQ19by33, SQ19by33, 121.0 / 361.0},
			QuadPoint{SQ19by33, -SQ19by33, SQ19by33, 121.0 / 361.0},
			QuadPoint{-SQ19by33, -SQ19by33, SQ19by33, 121.0 / 361.0},
			QuadPoint{SQ19by33, SQ19by33, -SQ19by33, 121.0 / 361.0},
			QuadPoint{-SQ19by33, SQ19by33, -SQ19by33, 121.0 / 361.0},
			QuadPoint{SQ19by33, -SQ19by33, -SQ19by33, 121.0 / 361.0},
			QuadPoint{-SQ19by33, -SQ19by33, -SQ19by33, 121.0 / 361.0},
		},
		27: []QuadPoint{
			QuadPoint{-0.774596669241483, -0.774596669241483, -0.774596669241483, 0.171467764060357},
			QuadPoint{+0.000000000000000, -0.774596669241483, -0.774596669241483, 0.274348422496571},
			QuadPoint{+0.774596669241483, -0.774596669241483, -0.774596669241483, 0.171467764060357},
			QuadPoint{-0.774596669241483, +0.000000000000000, -0.774596669241483, 0.274348422496571},
			QuadPoint{+0.000000000000000, +0.000000000000000, -0.774596669241483, 0.438957475994513},
			QuadPoint{+0.774596669241483, +0.000000000000000, -0.774596669241483, 0.274348422496571},
			QuadPoint{-0.774596669241483, +0.774596669241483, -0.774596669241483, 0.171467764060357},
			QuadPoint{+0.000000000000000, +0.774596669241483, -0.774596669241483, 0.274348422496571},
			QuadPoint{+0.774596669241483, +0.774596669241483, -0.774596669241483, 0.171467764060357},
			QuadPoint{-0.774596669241483, -0.774596669241483, +0.000000000000000, 0.274348422496571},
			QuadPoint{+0.000000000000000, -0.774596669241483, +0.000000000000000, 0.438957475994513},
			QuadPoint{+0.774596669241483, -0.774596669241483, +0.000000000000000, 0.274348422496571},
			QuadPoint{-0.774596669241483, +0.000000000000000, +0.000000000000000, 0.438957475994513},
			QuadPoint{+0.000000000000000, +0.000000000000000, +0.000000000000000, 0.702331961591221},
			QuadPoint{+0.774596669241483, +0.000000000000000, +0.000000000000000, 0.438957475994513},
			QuadPoint{-0.774596669241483, +0.774596669241483, +0.000000000000000, 0.274348422496571},
			QuadPoint{+0.000000000000000, +0.774596669241483, +0.000000000000000, 0.438957475994513},
			QuadPoint{+0.774596669241483, +0.774596669241483, +0.000000000000000, 0.274348422496571},
			QuadPoint{-0.774596669241483, -0.774596669241483, +0.774596669241483, 0.171467764060357},
			QuadPoint{+0.000000000000000, -0.774596669241483, +0.774596669241483, 0.274348422496571},
			QuadPoint{+0.774596669241483, -0.774596669241483, +0.774596669241483, 0.171467764060357},
			QuadPoint{-0.774596669241483, +0.000000000000000, +0.774596669241483, 0.274348422496571},
			QuadPoint{+0.000000000000000, +0.000000000000000, +0.774596669241483, 0.438957475994513},
			QuadPoint{+0.774596669241483, +0.000000000000000, +0.774596669241483, 0.274348422496571},
			QuadPoint{-0.774596669241483, +0.774596669241483, +0.774596669241483, 0.171467764060357},
			QuadPoint{+0.000000000000000, +0.774596669241483, +0.774596669241483, 0.274348422496571},
			QuadPoint{+0.774596669241483, +0.774596669241483, +0.774596669241483, 0.171467764060357},
		},
	}

	IntPointsOld["tri"] = map[int]QuadPoints{
		1: []QuadPoint{
			QuadPoint{1.0 / 3.0, 1.0 / 3.0, 0.0, 1.0 / 2.0},
		},

		3: []QuadPoint{
			QuadPoint{1.0 / 6.0, 1.0 / 6.0, 0.0, 1.0 / 6.0},
			QuadPoint{2.0 / 3.0, 1.0 / 6.0, 0.0, 1.0 / 6.0},
			QuadPoint{1.0 / 6.0, 2.0 / 3.0, 0.0, 1.0 / 6.0},
		},
		12: []QuadPoint{
			QuadPoint{0.873821971016996, 0.063089014491502, 0, 0.0254224531851035},
			QuadPoint{0.063089014491502, 0.873821971016996, 0, 0.0254224531851035},
			QuadPoint{0.063089014491502, 0.063089014491502, 0, 0.0254224531851035},
			QuadPoint{0.501426509658179, 0.249286745170910, 0, 0.0583931378631895},
			QuadPoint{0.249286745170910, 0.501426509658179, 0, 0.0583931378631895},
			QuadPoint{0.249286745170910, 0.249286745170910, 0, 0.0583931378631895},
			QuadPoint{0.053145049844817, 0.310352451033784, 0, 0.041425537809187},
			QuadPoint{0.310352451033784, 0.053145049844817, 0, 0.041425537809187},
			QuadPoint{0.053145049844817, 0.636502499121398, 0, 0.041425537809187},
			QuadPoint{0.310352451033784, 0.636502499121398, 0, 0.041425537809187},
			QuadPoint{0.636502499121398, 0.053145049844817, 0, 0.041425537809187},
			QuadPoint{0.636502499121398, 0.310352451033784, 0, 0.041425537809187},
		},
		16: []QuadPoint{
			QuadPoint{3.33333333333333E-01, 3.33333333333333E-01, 0.0, 7.21578038388935E-02},
			QuadPoint{8.14148234145540E-02, 4.59292588292723E-01, 0.0, 4.75458171336425E-02},
			QuadPoint{4.59292588292723E-01, 8.14148234145540E-02, 0.0, 4.75458171336425E-02},
			QuadPoint{4.59292588292723E-01, 4.59292588292723E-01, 0.0, 4.75458171336425E-02},
			QuadPoint{6.58861384496480E-01, 1.70569307751760E-01, 0.0, 5.16086852673590E-02},
			QuadPoint{1.70569307751760E-01, 6.58861384496480E-01, 0.0, 5.16086852673590E-02},
			QuadPoint{1.70569307751760E-01, 1.70569307751760E-01, 0.0, 5.16086852673590E-02},
			QuadPoint{8.98905543365938E-01, 5.05472283170310E-02, 0.0, 1.62292488115990E-02},
			QuadPoint{5.05472283170310E-02, 8.98905543365938E-01, 0.0, 1.62292488115990E-02},
			QuadPoint{5.05472283170310E-02, 5.05472283170310E-02, 0.0, 1.62292488115990E-02},
			QuadPoint{8.39477740995800E-03, 2.63112829634638E-01, 0.0, 1.36151570872175E-02},
			QuadPoint{7.28492392955404E-01, 8.39477740995800E-03, 0.0, 1.36151570872175E-02},
			QuadPoint{2.63112829634638E-01, 7.28492392955404E-01, 0.0, 1.36151570872175E-02},
			QuadPoint{8.39477740995800E-03, 7.28492392955404E-01, 0.0, 1.36151570872175E-02},
			QuadPoint{7.28492392955404E-01, 2.63112829634638E-01, 0.0, 1.36151570872175E-02},
			QuadPoint{2.63112829634638E-01, 8.39477740995800E-03, 0.0, 1.36151570872175E-02},
		},
	}

	IntPointsOld["tet"] = map[int]QuadPoints{
		1: []QuadPoint{
			QuadPoint{1.0 / 4.0, 1.0 / 4.0, 1.0 / 4.0, 1.0 / 6.0},
		},
		4: []QuadPoint{
			QuadPoint{(5.0 + 3.0*utl.SQ5) / 20.0, (5.0 - utl.SQ5) / 20.0, (5.0 - utl.SQ5) / 20.0, 1.0 / 24},
			QuadPoint{(5.0 - utl.SQ5) / 20.0, (5.0 + 3.0*utl.SQ5) / 20.0, (5.0 - utl.SQ5) / 20.0, 1.0 / 24},
			QuadPoint{(5.0 - utl.SQ5) / 20.0, (5.0 - utl.SQ5) / 20.0, (5.0 + 3.0*utl.SQ5) / 20.0, 1.0 / 24},
			QuadPoint{(5.0 - utl.SQ5) / 20.0, (5.0 - utl.SQ5) / 20.0, (5.0 - utl.SQ5) / 20.0, 1.0 / 24},
		},
		5: []QuadPoint{
			QuadPoint{1.0 / 4.0, 1.0 / 4.0, 1.0 / 4.0, -2.0 / 15.0},
			QuadPoint{1.0 / 6.0, 1.0 / 6.0, 1.0 / 6.0, 3.0 / 40.0},
			QuadPoint{1.0 / 6.0, 1.0 / 6.0, 1.0 / 2.0, 3.0 / 40.0},
			QuadPoint{1.0 / 6.0, 1.0 / 2.0, 1.0 / 6.0, 3.0 / 40.0},
			QuadPoint{1.0 / 2.0, 1.0 / 6.0, 1.0 / 6.0, 3.0 / 40.0},
		},
		6: []QuadPoint{
			QuadPoint{+1.0, +0.0, +0.0, 4.0 / 3.0},
			QuadPoint{-1.0, +0.0, +0.0, 4.0 / 3.0},
			QuadPoint{+0.0, +1.0, +0.0, 4.0 / 3.0},
			QuadPoint{+0.0, -1.0, +0.0, 4.0 / 3.0},
			QuadPoint{+0.0, +0.0, +1.0, 4.0 / 3.0},
			QuadPoint{+0.0, +0.0, -1.0, 4.0 / 3.0},
		},
	}
}
