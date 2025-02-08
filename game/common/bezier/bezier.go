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
}

// Curve implements Bezier curve calculation according to the algorithm of Robert D. Miller.
//
// Graphics Gems 5, 'Quick and Simple Bézier Curve Drawing', pages 206-209.
type Curve []point

// NewCurve returns a Curve initialized with the control points in cp.
func New(cp []*geom.Vector2) Curve {
	if len(cp) == 0 {
		return nil
	}
	c := make(Curve, len(cp))
	for i, p := range cp {
		c[i].Point = *p
	}

	var w float64
	for i, p := range c {
		switch i {
		case 0:
			w = 1
		case 1:
			w = float64(len(c)) - 1
		default:
			w *= float64(len(c)-i) / float64(i)
		}
		c[i].Control.X = p.Point.X * w
		c[i].Control.Y = p.Point.Y * w
	}

	return c
}

// Point returns the point at t along the curve, where 0 ≤ t ≤ 1.
func (c Curve) Point(t float64) *geom.Vector2 {
	c[0].Point = c[0].Control
	u := t
	for i, p := range c[1:] {
		c[i+1].Point = geom.Vector2{
			X: p.Control.X * float64(u),
			Y: p.Control.Y * float64(u),
		}
		u *= t
	}

	var (
		t1 = 1 - t
		tt = t1
	)
	p := c[len(c)-1].Point
	for i := len(c) - 2; i >= 0; i-- {
		p.X += c[i].Point.X * float64(tt)
		p.Y += c[i].Point.Y * float64(tt)
		tt *= t1
	}

	return &p
}

// Curve returns a slice of vg.Point, p, filled with points along the Bézier curve described by c.
// If the length of p is less than 2, the curve points are undefined. The length of p is not
// altered by the call.
func (c Curve) Curve(p []*geom.Vector2) []*geom.Vector2 {
	for i, nf := 0, float64(len(p)-1); i < len(p); i++ {
		p[i] = c.Point(float64(i) / nf)
	}
	return p
}
