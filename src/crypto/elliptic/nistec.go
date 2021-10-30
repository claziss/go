// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elliptic

import (
	"crypto/elliptic/internal/nistec"
	"crypto/rand"
	"math/big"
)

var p224 = &nistCurve[*nistec.P224Point]{
	newPoint:     nistec.NewP224Point,
	newGenerator: nistec.NewP224Generator,
}

func initP224() {
	p224.params = &CurveParams{
		Name:    "P-224",
		BitSize: 224,
		// FIPS 186-4, section D.1.2.2
		P:  bigFromDecimal("26959946667150639794667015087019630673557916260026308143510066298881"),
		N:  bigFromDecimal("26959946667150639794667015087019625940457807714424391721682722368061"),
		B:  bigFromHex("b4050a850c04b3abf54132565044b0b7d7bfd8ba270b39432355ffb4"),
		Gx: bigFromHex("b70e0cbd6bb4bf7f321390b94a03c1d356c21122343280d6115c1d21"),
		Gy: bigFromHex("bd376388b5f723fb4c22dfe6cd4375a05a07476444d5819985007e34"),
	}
}

var p384 = &nistCurve[*nistec.P384Point]{
	newPoint:     nistec.NewP384Point,
	newGenerator: nistec.NewP384Generator,
}

func initP384() {
	p384.params = &CurveParams{
		Name:    "P-384",
		BitSize: 384,
		// FIPS 186-4, section D.1.2.4
		P: bigFromDecimal("394020061963944792122790401001436138050797392704654" +
			"46667948293404245721771496870329047266088258938001861606973112319"),
		N: bigFromDecimal("394020061963944792122790401001436138050797392704654" +
			"46667946905279627659399113263569398956308152294913554433653942643"),
		B: bigFromHex("b3312fa7e23ee7e4988e056be3f82d19181d9c6efe8141120314088" +
			"f5013875ac656398d8a2ed19d2a85c8edd3ec2aef"),
		Gx: bigFromHex("aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741" +
			"e082542a385502f25dbf55296c3a545e3872760ab7"),
		Gy: bigFromHex("3617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da31" +
			"13b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f"),
	}
}

var p521 = &nistCurve[*nistec.P521Point]{
	newPoint:     nistec.NewP521Point,
	newGenerator: nistec.NewP521Generator,
}

func initP521() {
	p521.params = &CurveParams{
		Name:    "P-521",
		BitSize: 521,
		// FIPS 186-4, section D.1.2.5
		P: bigFromDecimal("68647976601306097149819007990813932172694353001433" +
			"0540939446345918554318339765605212255964066145455497729631139148" +
			"0858037121987999716643812574028291115057151"),
		N: bigFromDecimal("68647976601306097149819007990813932172694353001433" +
			"0540939446345918554318339765539424505774633321719753296399637136" +
			"3321113864768612440380340372808892707005449"),
		B: bigFromHex("0051953eb9618e1c9a1f929a21a0b68540eea2da725b99b315f3b8" +
			"b489918ef109e156193951ec7e937b1652c0bd3bb1bf073573df883d2c34f1ef" +
			"451fd46b503f00"),
		Gx: bigFromHex("00c6858e06b70404e9cd9e3ecb662395b4429c648139053fb521f8" +
			"28af606b4d3dbaa14b5e77efe75928fe1dc127a2ffa8de3348b3c1856a429bf9" +
			"7e7e31c2e5bd66"),
		Gy: bigFromHex("011839296a789a3bc0045c8a5fb42c7d1bd998f54449579b446817" +
			"afbd17273e662c97ee72995ef42640c550b9013fad0761353c7086a272c24088" +
			"be94769fd16650"),
	}
}

// nistCurve is a Curve implementation based on a nistec Point.
//
// It's a wrapper that exposes the big.Int-based Curve interface and encodes the
// legacy idiosyncrasies it requires, such as invalid and infinity point
// handling.
//
// To interact with the nistec package, points are encoded into and decoded from
// properly formatted byte slices. All big.Int use is limited to this package.
// Encoding and decoding is 1/1000th of the runtime of a scalar multiplication,
// so the overhead is acceptable.
type nistCurve[Point nistPoint[Point]] struct {
	newPoint     func() Point
	newGenerator func() Point
	params       *CurveParams
}

// nistPoint is a generic constraint for the nistec Point types.
type nistPoint[T any] interface {
	Bytes() []byte
	SetBytes([]byte) (T, error)
	Add(T, T) T
	Double(T) T
	ScalarMult(T, []byte) T
}

func (curve *nistCurve[Point]) Params() *CurveParams {
	return curve.params
}

func (curve *nistCurve[Point]) IsOnCurve(x, y *big.Int) bool {
	// IsOnCurve is documented to reject (0, 0), the conventional point at
	// infinity, which however is accepted by pointFromAffine.
	if x.Sign() == 0 && y.Sign() == 0 {
		return false
	}
	_, ok := curve.pointFromAffine(x, y)
	return ok
}

func (curve *nistCurve[Point]) pointFromAffine(x, y *big.Int) (p Point, ok bool) {
	// (0, 0) is by convention the point at infinity, which can't be represented
	// in affine coordinates. Marshal incorrectly encodes it as an uncompressed
	// point, which SetBytes would correctly reject. See Issue 37294.
	if x.Sign() == 0 && y.Sign() == 0 {
		return curve.newPoint(), true
	}
	if x.Sign() < 0 || y.Sign() < 0 {
		return curve.newPoint(), false
	}
	if x.BitLen() > curve.params.BitSize || y.BitLen() > curve.params.BitSize {
		return *new(Point), false
	}
	p, err := curve.newPoint().SetBytes(Marshal(curve, x, y))
	if err != nil {
		return *new(Point), false
	}
	return p, true
}

func (curve *nistCurve[Point]) pointToAffine(p Point) (x, y *big.Int) {
	out := p.Bytes()
	if len(out) == 1 && out[0] == 0 {
		// This is the correct encoding of the point at infinity, which
		// Unmarshal does not support. See Issue 37294.
		return new(big.Int), new(big.Int)
	}
	x, y = Unmarshal(curve, out)
	if x == nil {
		panic("crypto/elliptic: internal error: Unmarshal rejected a valid point encoding")
	}
	return x, y
}

// randomPoint returns a random point on the curve. It's used when Add,
// Double, or ScalarMult are fed a point not on the curve, which is undefined
// behavior. Originally, we used to do the math on it anyway (which allows
// invalid curve attacks) and relied on the caller and Unmarshal to avoid this
// happening in the first place. Now, we just can't construct a nistec Point
// for an invalid pair of coordinates, because that API is safer. If we panic,
// we risk introducing a DoS. If we return nil, we risk a panic. If we return
// the input, ecdsa.Verify might fail open. The safest course seems to be to
// return a valid, random point, which hopefully won't help the attacker.
func (curve *nistCurve[Point]) randomPoint() (x, y *big.Int) {
	_, x, y, err := GenerateKey(curve, rand.Reader)
	if err != nil {
		panic("crypto/elliptic: failed to generate random point")
	}
	return x, y
}

func (curve *nistCurve[Point]) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	p1, ok := curve.pointFromAffine(x1, y1)
	if !ok {
		return curve.randomPoint()
	}
	p2, ok := curve.pointFromAffine(x2, y2)
	if !ok {
		return curve.randomPoint()
	}
	return curve.pointToAffine(p1.Add(p1, p2))
}

func (curve *nistCurve[Point]) Double(x1, y1 *big.Int) (*big.Int, *big.Int) {
	p, ok := curve.pointFromAffine(x1, y1)
	if !ok {
		return curve.randomPoint()
	}
	return curve.pointToAffine(p.Double(p))
}

func (curve *nistCurve[Point]) ScalarMult(Bx, By *big.Int, scalar []byte) (*big.Int, *big.Int) {
	p, ok := curve.pointFromAffine(Bx, By)
	if !ok {
		return curve.randomPoint()
	}
	return curve.pointToAffine(p.ScalarMult(p, scalar))
}

func (curve *nistCurve[Point]) ScalarBaseMult(scalar []byte) (*big.Int, *big.Int) {
	p := curve.newGenerator()
	return curve.pointToAffine(p.ScalarMult(p, scalar))
}

func bigFromDecimal(s string) *big.Int {
	b, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid encoding")
	}
	return b
}

func bigFromHex(s string) *big.Int {
	b, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic("invalid encoding")
	}
	return b
}