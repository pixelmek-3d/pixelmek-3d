// Based on https://github.com/gonum/plot/blob/master/tools/bezier/bezier.go
// with the following BSD-3-Clause license:
//
// Copyright ©2013 The Gonum Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//     * Redistributions of source code must retain the above copyright
//       notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above copyright
//       notice, this list of conditions and the following disclaimer in the
//       documentation and/or other materials provided with the distribution.
//     * Neither the name of the Gonum project nor the names of its authors and
//       contributors may be used to endorse or promote products derived from this
//       software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package bezier

import (
	"github.com/harbdog/raycaster-go/geom"
)

type point struct {
	Point, Control geom.Vector2
	Weighting      float64
}

// Curve implements Bezier curve calculation according to the algorithm of Robert D. Miller.
//
// Graphics Gems 5, 'Quick and Simple Bézier Curve Drawing', pages 206-209.
type Curve struct {
	p []point
	w []float64
}

// NewCurve returns a Curve initialized with the control points in cp.
func New(cp []*geom.Vector2, weight float64) *Curve {
	if len(cp) == 0 {
		return nil
	}
	c := &Curve{p: make([]point, len(cp)), w: make([]float64, len(cp))}
	for i, p := range cp {
		c.p[i].Point = *p

		// TODO: if needed later, could support a different weight value at each control point
		c.w[i] = weight
	}

	var s float64
	for i, p := range c.p {
		switch i {
		case 0:
			s = 1
		case 1:
			s = float64(len(c.p)) - 1
		default:
			s *= float64(len(c.p)-i) / float64(i)
		}
		w := c.w[i]

		c.p[i].Control.X = p.Point.X * w * s
		c.p[i].Control.Y = p.Point.Y * w * s

		c.p[i].Weighting = w * s
	}

	return c
}

// Point returns the point at t along the curve, where 0 ≤ t ≤ 1.
func (c Curve) Point(t float64) *geom.Vector2 {
	b := Curve{p: make([]point, len(c.p))}

	b.p[0].Point = c.p[0].Control
	b.p[0].Weighting = c.p[0].Weighting
	u := t
	for i, p := range c.p[1:] {
		b.p[i+1].Point = geom.Vector2{
			X: p.Control.X * u,
			Y: p.Control.Y * u,
		}
		b.p[i+1].Weighting = p.Weighting * u

		u *= t
	}

	var (
		t1 = 1 - t
		tt = t1
	)
	p := b.p[len(b.p)-1].Point
	wp := b.p[len(b.p)-1].Weighting
	for i := len(b.p) - 2; i >= 0; i-- {
		p.X += b.p[i].Point.X * tt
		p.Y += b.p[i].Point.Y * tt
		wp += b.p[i].Weighting * tt

		tt *= t1
	}

	p.X /= wp
	p.Y /= wp

	return &p
}

// Curve returns a slice of points, p, filled with points along the Bézier curve described by c.
// If the length of p is less than 2, the curve points are undefined. The length of p is not
// altered by the call.
func (c Curve) Curve(p []*geom.Vector2) []*geom.Vector2 {
	for i, nf := 0, float64(len(p)-1); i < len(p); i++ {
		p[i] = c.Point(float64(i) / nf)
	}
	return p
}
