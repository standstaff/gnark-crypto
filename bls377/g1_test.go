// Copyright 2020 ConsenSys AG
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gurvy DO NOT EDIT

package bls377

import (
	"fmt"
	"math/big"
	"runtime"
	"testing"

	"github.com/consensys/gurvy/bls377/fp"
	"github.com/consensys/gurvy/bls377/fr"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// ------------------------------------------------------------
// utils
func fuzzJacobianG1(p *G1Jac, f fp.Element) G1Jac {
	var res G1Jac
	res.X.Mul(&p.X, &f).Mul(&res.X, &f)
	res.Y.Mul(&p.Y, &f).Mul(&res.Y, &f).Mul(&res.Y, &f)
	res.Z.Mul(&p.Z, &f)
	return res
}

func fuzzProjectiveG1(p *G1Proj, f fp.Element) G1Proj {
	var res G1Proj
	res.X.Mul(&p.X, &f)
	res.Y.Mul(&p.Y, &f)
	res.Z.Mul(&p.Z, &f)
	return res
}

func fuzzExtendedJacobianG1(p *g1JacExtended, f fp.Element) g1JacExtended {
	var res g1JacExtended
	var ff, fff fp.Element
	ff.Square(&f)
	fff.Mul(&ff, &f)
	res.X.Mul(&p.X, &ff)
	res.Y.Mul(&p.Y, &fff)
	res.ZZ.Mul(&p.ZZ, &ff)
	res.ZZZ.Mul(&p.ZZZ, &fff)
	return res
}

// ------------------------------------------------------------
// tests

func TestG1IsOnCurve(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenFp()
	properties.Property("g1Gen (affine) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2 G1Affine
			op1.FromJacobian(&g1Gen)
			op2.FromJacobian(&g1Gen)
			op2.Y.Mul(&op2.Y, &a)
			return op1.IsOnCurve() && !op2.IsOnCurve()
		},
		genFuzz1,
	))

	properties.Property("g1Gen (Jacobian) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2, op3 G1Jac
			op1.Set(&g1Gen)
			op3.Set(&g1Gen)

			op2 = fuzzJacobianG1(&g1Gen, a)
			op3.Y.Mul(&op3.Y, &a)
			return op1.IsOnCurve() && op2.IsOnCurve() && !op3.IsOnCurve()
		},
		genFuzz1,
	))

	properties.Property("g1Gen (projective) should be on the curve", prop.ForAll(
		func(a fp.Element) bool {
			var op1, op2, op3 G1Proj
			op1.FromJacobian(&g1Gen)
			op2.FromJacobian(&g1Gen)
			op3.FromJacobian(&g1Gen)

			op2 = fuzzProjectiveG1(&op1, a)
			op3.Y.Mul(&op3.Y, &a)
			return op1.IsOnCurve() && op2.IsOnCurve() && !op3.IsOnCurve()
		},
		genFuzz1,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1Conversions(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenFp()
	genFuzz2 := GenFp()

	properties.Property("Affine representation should be independent of the Jacobian representative", prop.ForAll(
		func(a fp.Element) bool {
			g := fuzzJacobianG1(&g1Gen, a)
			var op1 G1Affine
			op1.FromJacobian(&g)
			return op1.X.Equal(&g1Gen.X) && op1.Y.Equal(&g1Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Affine representation should be independent of a Extended Jacobian representative", prop.ForAll(
		func(a fp.Element) bool {
			var g g1JacExtended
			g.X.Set(&g1Gen.X)
			g.Y.Set(&g1Gen.Y)
			g.ZZ.Set(&g1Gen.Z)
			g.ZZZ.Set(&g1Gen.Z)
			gfuzz := fuzzExtendedJacobianG1(&g, a)

			var op1 G1Affine
			gfuzz.ToAffine(&op1)
			return op1.X.Equal(&g1Gen.X) && op1.Y.Equal(&g1Gen.Y)
		},
		genFuzz1,
	))

	properties.Property("Projective representation should be independent of a Jacobian representative", prop.ForAll(
		func(a fp.Element) bool {

			g := fuzzJacobianG1(&g1Gen, a)

			var op1 G1Proj
			op1.FromJacobian(&g)
			var u, v fp.Element
			u.Mul(&g.X, &g.Z)
			v.Square(&g.Z).Mul(&v, &g.Z)

			return op1.X.Equal(&u) && op1.Y.Equal(&g.Y) && op1.Z.Equal(&v)
		},
		genFuzz1,
	))

	properties.Property("Jacobian representation should be the same as the affine representative", prop.ForAll(
		func(a fp.Element) bool {
			var g G1Jac
			var op1 G1Affine
			op1.X.Set(&g1Gen.X)
			op1.Y.Set(&g1Gen.Y)

			var one fp.Element
			one.SetOne()

			g.FromAffine(&op1)

			return g.X.Equal(&g1Gen.X) && g.Y.Equal(&g1Gen.Y) && g.Z.Equal(&one)
		},
		genFuzz1,
	))

	properties.Property("Converting affine symbol for infinity to Jacobian should output correct infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G1Affine
			g.X.SetZero()
			g.Y.SetZero()
			var op1 G1Jac
			op1.FromAffine(&g)
			var one, zero fp.Element
			one.SetOne()
			return op1.X.Equal(&one) && op1.Y.Equal(&one) && op1.Z.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to affine should output infinity symbol in Affine", prop.ForAll(
		func() bool {
			var g G1Affine
			var op1 g1JacExtended
			var zero fp.Element
			op1.X.Set(&g1Gen.X)
			op1.Y.Set(&g1Gen.Y)
			op1.ToAffine(&g)
			return g.X.Equal(&zero) && g.Y.Equal(&zero)
		},
	))

	properties.Property("Converting infinity in extended Jacobian to Jacobian should output infinity in Jacobian", prop.ForAll(
		func() bool {
			var g G1Jac
			var op1 g1JacExtended
			var zero, one fp.Element
			one.SetOne()
			op1.X.Set(&g1Gen.X)
			op1.Y.Set(&g1Gen.Y)
			op1.ToJac(&g)
			return g.X.Equal(&one) && g.Y.Equal(&one) && g.Z.Equal(&zero)
		},
	))

	properties.Property("[Jacobian] Two representatives of the same class should be equal", prop.ForAll(
		func(a, b fp.Element) bool {
			op1 := fuzzJacobianG1(&g1Gen, a)
			op2 := fuzzJacobianG1(&g1Gen, b)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1Ops(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)
	genFuzz1 := GenFp()
	genFuzz2 := GenFp()

	genScalar := GenFr()

	properties.Property("[Jacobian] Add should call double when having adding the same point", prop.ForAll(
		func(a, b fp.Element) bool {
			fop1 := fuzzJacobianG1(&g1Gen, a)
			fop2 := fuzzJacobianG1(&g1Gen, b)
			var op1, op2 G1Jac
			op1.Set(&fop1).AddAssign(&fop2)
			op2.Double(&fop2)
			return op1.Equal(&op2)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the opposite of a point to itself should output inf", prop.ForAll(
		func(a, b fp.Element) bool {
			fop1 := fuzzJacobianG1(&g1Gen, a)
			fop2 := fuzzJacobianG1(&g1Gen, b)
			fop2.Neg(&fop2)
			fop1.AddAssign(&fop2)
			return fop1.Equal(&g1Infinity)
		},
		genFuzz1,
		genFuzz2,
	))

	properties.Property("[Jacobian] Adding the inf to a point should not modify the point", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzJacobianG1(&g1Gen, a)
			fop1.AddAssign(&g1Infinity)
			var op2 G1Jac
			op2.Set(&g1Infinity)
			op2.AddAssign(&g1Gen)
			return fop1.Equal(&g1Gen) && op2.Equal(&g1Gen)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian Extended] mAdd (-G) should equal mSub(G)", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzJacobianG1(&g1Gen, a)
			var p1, p1Neg G1Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g1JacExtended
			o1.mAdd(&p1Neg)
			o2.mSub(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian Extended] double (-G) should equal doubleNeg(G)", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzJacobianG1(&g1Gen, a)
			var p1, p1Neg G1Affine
			p1.FromJacobian(&fop1)
			p1Neg = p1
			p1Neg.Y.Neg(&p1Neg.Y)
			var o1, o2 g1JacExtended
			o1.double(&p1Neg)
			o2.doubleNeg(&p1)

			return o1.X.Equal(&o2.X) &&
				o1.Y.Equal(&o2.Y) &&
				o1.ZZ.Equal(&o2.ZZ) &&
				o1.ZZZ.Equal(&o2.ZZZ)
		},
		genFuzz1,
	))

	properties.Property("[Jacobian] Addmix the negation to itself should output 0", prop.ForAll(
		func(a fp.Element) bool {
			fop1 := fuzzJacobianG1(&g1Gen, a)
			fop1.Neg(&fop1)
			var op2 G1Affine
			op2.FromJacobian(&g1Gen)
			fop1.AddMixed(&op2)
			return fop1.Equal(&g1Infinity)
		},
		genFuzz1,
	))

	properties.Property("scalar multiplication (double and add) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G1Jac
			var gaff G1Affine
			gaff.FromJacobian(&g1Gen)
			g.ScalarMultiplication(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg G1Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMultiplication(&gaff, &rminusone)
			gneg.Neg(&g1Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMultiplication(&gaff, &scalar)
			op2.ScalarMultiplication(&gaff, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&g1Infinity) && !op1.Equal(&g1Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("scalar multiplication (GLV) should depend only on the scalar mod r", prop.ForAll(
		func(s fr.Element) bool {

			r := fr.Modulus()
			var g G1Jac
			var gaff G1Affine
			gaff.FromJacobian(&g1Gen)
			g.ScalarMulGLV(&gaff, r)

			var scalar, blindedScalard, rminusone big.Int
			var op1, op2, op3, gneg G1Jac
			rminusone.SetUint64(1).Sub(r, &rminusone)
			op3.ScalarMulGLV(&gaff, &rminusone)
			gneg.Neg(&g1Gen)
			s.ToBigIntRegular(&scalar)
			blindedScalard.Add(&scalar, r)
			op1.ScalarMulGLV(&gaff, &scalar)
			op2.ScalarMulGLV(&gaff, &blindedScalard)

			return op1.Equal(&op2) && g.Equal(&g1Infinity) && !op1.Equal(&g1Infinity) && gneg.Equal(&op3)

		},
		genScalar,
	))

	properties.Property("GLV and Double and Add should output the same result", prop.ForAll(
		func(s fr.Element) bool {

			var r big.Int
			var op1, op2 G1Jac
			var gaff G1Affine
			s.ToBigIntRegular(&r)
			gaff.FromJacobian(&g1Gen)
			op1.ScalarMultiplication(&gaff, &r)
			op2.ScalarMulGLV(&gaff, &r)
			return op1.Equal(&op2) && !op1.Equal(&g1Infinity)

		},
		genScalar,
	))

	// note : this test is here as we expect to have a different multiExp than the above bucket method
	// for small number of points
	properties.Property("Multi exponentation (<50points) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var g G1Jac
			g.Set(&g1Gen)

			// mixer ensures that all the words of a fpElement are set
			samplePoints := make([]G1Affine, 30)
			sampleScalars := make([]fr.Element, 30)

			for i := 1; i <= 30; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
				samplePoints[i-1].FromJacobian(&g)
				g.AddAssign(&g1Gen)
			}

			var op1MultiExp G1Jac
			op1MultiExp.MultiExp(samplePoints, sampleScalars)

			var finalBigScalar fr.Element
			var finalBigScalarBi big.Int
			var op1ScalarMul G1Jac
			var op1Aff G1Affine
			op1Aff.FromJacobian(&g1Gen)
			finalBigScalar.SetString("9455").MulAssign(&mixer)
			finalBigScalar.ToBigIntRegular(&finalBigScalarBi)
			op1ScalarMul.ScalarMultiplication(&op1Aff, &finalBigScalarBi)

			return op1ScalarMul.Equal(&op1MultiExp)
		},
		genScalar,
	))
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1MultiExp(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	// size of the multiExps
	const nbSamples = 500

	// multi exp points
	var samplePoints [nbSamples]G1Affine
	var g G1Jac
	g.Set(&g1Gen)
	for i := 1; i <= nbSamples; i++ {
		samplePoints[i-1].FromJacobian(&g)
		g.AddAssign(&g1Gen)
	}

	// final scalar to use in double and add method (without mixer factor)
	// n(n+1)(2n+1)/6  (sum of the squares from 1 to n)
	var scalar big.Int
	scalar.SetInt64(nbSamples)
	scalar.Mul(&scalar, new(big.Int).SetInt64(nbSamples+1))
	scalar.Mul(&scalar, new(big.Int).SetInt64(2*nbSamples+1))
	scalar.Div(&scalar, new(big.Int).SetInt64(6))

	properties.Property("Multi exponentation (c=4) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 4)
			result.msmC4(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=5) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 5)
			result.msmC5(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=6) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 6)
			result.msmC6(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=7) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 7)
			result.msmC7(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=8) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 8)
			result.msmC8(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=9) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 9)
			result.msmC9(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=10) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 10)
			result.msmC10(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=11) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 11)
			result.msmC11(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=12) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 12)
			result.msmC12(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=13) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 13)
			result.msmC13(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=14) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 14)
			result.msmC14(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	properties.Property("Multi exponentation (c=15) should be consistant with sum of square", prop.ForAll(
		func(mixer fr.Element) bool {

			var result, expected G1Jac

			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			// semaphore to limit number of cpus
			opt := NewMultiExpOptions(runtime.NumCPU())
			opt.lock.Lock()
			scalars := partitionScalars(sampleScalars[:], 15)
			result.msmC15(samplePoints[:], scalars, opt)

			// compute expected result with double and add
			var finalScalar, mixerBigInt big.Int
			finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
			expected.ScalarMultiplication(&g1GenAff, &finalScalar)

			return result.Equal(&expected)
		},
		genScalar,
	))

	if !testing.Short() {

		properties.Property("Multi exponentation (c=16) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G1Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 16)
				result.msmC16(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g1GenAff, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	if !testing.Short() {

		properties.Property("Multi exponentation (c=20) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G1Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 20)
				result.msmC20(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g1GenAff, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	if !testing.Short() {

		properties.Property("Multi exponentation (c=21) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G1Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 21)
				result.msmC21(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g1GenAff, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	if !testing.Short() {

		properties.Property("Multi exponentation (c=22) should be consistant with sum of square", prop.ForAll(
			func(mixer fr.Element) bool {

				var result, expected G1Jac

				// mixer ensures that all the words of a fpElement are set
				var sampleScalars [nbSamples]fr.Element

				for i := 1; i <= nbSamples; i++ {
					sampleScalars[i-1].SetUint64(uint64(i)).
						MulAssign(&mixer).
						FromMont()
				}

				// semaphore to limit number of cpus
				opt := NewMultiExpOptions(runtime.NumCPU())
				opt.lock.Lock()
				scalars := partitionScalars(sampleScalars[:], 22)
				result.msmC22(samplePoints[:], scalars, opt)

				// compute expected result with double and add
				var finalScalar, mixerBigInt big.Int
				finalScalar.Mul(&scalar, mixer.ToBigIntRegular(&mixerBigInt))
				expected.ScalarMultiplication(&g1GenAff, &finalScalar)

				return result.Equal(&expected)
			},
			genScalar,
		))

	}

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

func TestG1BatchScalarMultiplication(t *testing.T) {

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10

	properties := gopter.NewProperties(parameters)

	genScalar := GenFr()

	// size of the multiExps
	const nbSamples = 500

	properties.Property("BatchScalarMultiplication should be consistant with individual scalar multiplications", prop.ForAll(
		func(mixer fr.Element) bool {
			// mixer ensures that all the words of a fpElement are set
			var sampleScalars [nbSamples]fr.Element

			for i := 1; i <= nbSamples; i++ {
				sampleScalars[i-1].SetUint64(uint64(i)).
					MulAssign(&mixer).
					FromMont()
			}

			result := BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:])

			if len(result) != len(sampleScalars) {
				return false
			}

			for i := 0; i < len(result); i++ {
				var expectedJac G1Jac
				var expected G1Affine
				var b big.Int
				expectedJac.ScalarMulGLV(&g1GenAff, sampleScalars[i].ToBigInt(&b))
				expected.FromJacobian(&expectedJac)
				if !result[i].Equal(&expected) {
					return false
				}
			}
			return true
		},
		genScalar,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// ------------------------------------------------------------
// benches

func BenchmarkG1BatchScalarMul(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 15
	const nbSamples = 1 << pow

	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
	}

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				_ = BatchScalarMultiplicationG1(&g1GenAff, sampleScalars[:using])
			}
		})
	}
}

func BenchmarkG1ScalarMul(b *testing.B) {

	var scalar big.Int
	scalar.SetString("5243587517512619047944770508185965837690552500527637822603658699938581184513", 10)

	var doubleAndAdd G1Jac

	b.Run("double and add", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			doubleAndAdd.ScalarMultiplication(&g1GenAff, &scalar)
		}
	})

	var glv G1Jac
	b.Run("GLV", func(b *testing.B) {
		b.ResetTimer()
		for j := 0; j < b.N; j++ {
			glv.ScalarMulGLV(&g1GenAff, &scalar)
		}
	})

}

func BenchmarkG1Add(b *testing.B) {
	var a G1Jac
	a.Double(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddAssign(&g1Gen)
	}
}

func BenchmarkG1mAdd(b *testing.B) {
	var a g1JacExtended
	a.double(&g1GenAff)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.mAdd(&c)
	}

}

func BenchmarkG1AddMixed(b *testing.B) {
	var a G1Jac
	a.Double(&g1Gen)

	var c G1Affine
	c.FromJacobian(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddMixed(&c)
	}

}

func BenchmarkG1Double(b *testing.B) {
	var a G1Jac
	a.Set(&g1Gen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.DoubleAssign()
	}

}

func BenchmarkG1MultiExpLargeG1(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 27
	const nbSamples = 1 << pow

	var samplePoints [nbSamples]G1Affine
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1] = g1GenAff
	}

	var testPoint G1Jac

	for i := 23; i <= pow; i++ {
		for c := 16; c <= 22; c++ {
			for cpus := 2; cpus <= 8; cpus *= 2 {
				using := 1 << i

				opt := NewMultiExpOptions(cpus)
				opt.C = uint64(c)
				b.Run(fmt.Sprintf("%d points, c = %d, cpus = %d", using, c, cpus), func(b *testing.B) {
					b.ResetTimer()
					for j := 0; j < b.N; j++ {
						testPoint.MultiExp(samplePoints[:using], sampleScalars[:using], opt)
					}
				})
			}
		}
	}

}

func BenchmarkG1MultiExpG1(b *testing.B) {
	// ensure every words of the scalars are filled
	var mixer fr.Element
	mixer.SetString("7716837800905789770901243404444209691916730933998574719964609384059111546487")

	const pow = 24
	const nbSamples = 1 << pow

	var samplePoints [nbSamples]G1Affine
	var sampleScalars [nbSamples]fr.Element

	for i := 1; i <= nbSamples; i++ {
		sampleScalars[i-1].SetUint64(uint64(i)).
			Mul(&sampleScalars[i-1], &mixer).
			FromMont()
		samplePoints[i-1] = g1GenAff
	}

	var testPoint G1Jac

	for i := 5; i <= pow; i++ {
		using := 1 << i

		b.Run(fmt.Sprintf("%d points", using), func(b *testing.B) {
			b.ResetTimer()
			for j := 0; j < b.N; j++ {
				testPoint.MultiExp(samplePoints[:using], sampleScalars[:using])
			}
		})
	}
}
