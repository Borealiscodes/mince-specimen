// Package felix provides overflow-checked exact p/q rational arithmetic
// for bounded-gas capability budgets and compression telemetry.
// Zero third-party dependencies (Go standard library only, zero floats).
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package felix

import (
	"errors"
	"math/bits"
)

var (
	ErrZeroDenominator  = errors.New("felix: rational denominator cannot be zero")
	ErrRationalOverflow = errors.New("felix: 64-bit integer overflow in rational operation")
)

// Rational represents an exact reduced fraction P/Q with Q > 0.
type Rational struct {
	P uint64
	Q uint64
}

// NewRational constructs a reduced P/Q rational, failing closed if Q == 0.
func NewRational(p, q uint64) (Rational, error) {
	if q == 0 {
		return Rational{}, ErrZeroDenominator
	}
	if p == 0 {
		return Rational{P: 0, Q: 1}, nil
	}
	g := gcd(p, q)
	return Rational{P: p / g, Q: q / g}, nil
}

// Add computes (a.P/a.Q) + (b.P/b.Q) with strict 128-bit intermediate overflow checking.
func (a Rational) Add(b Rational) (Rational, error) {
	if a.Q == 0 || b.Q == 0 {
		return Rational{}, ErrZeroDenominator
	}
	// Cross-multiply into 128-bit intermediates to prevent silent wrap-around
	hi1, lo1 := bits.Mul64(a.P, b.Q)
	hi2, lo2 := bits.Mul64(b.P, a.Q)

	sumLo, carry := bits.Add64(lo1, lo2, 0)
	sumHi, ovf := bits.Add64(hi1, hi2, carry)
	if ovf != 0 || sumHi != 0 {
		return Rational{}, ErrRationalOverflow
	}

	denHi, denLo := bits.Mul64(a.Q, b.Q)
	if denHi != 0 {
		return Rational{}, ErrRationalOverflow
	}

	return NewRational(sumLo, denLo)
}

// Mul computes (a.P/a.Q) * (b.P/b.Q) with pre-reduction to minimize overflow risk.
func (a Rational) Mul(b Rational) (Rational, error) {
	if a.Q == 0 || b.Q == 0 {
		return Rational{}, ErrZeroDenominator
	}
	g1 := gcd(a.P, b.Q)
	g2 := gcd(b.P, a.Q)

	numHi, numLo := bits.Mul64(a.P/g1, b.P/g2)
	if numHi != 0 {
		return Rational{}, ErrRationalOverflow
	}
	denHi, denLo := bits.Mul64(a.Q/g2, b.Q/g1)
	if denHi != 0 {
		return Rational{}, ErrRationalOverflow
	}

	return NewRational(numLo, denLo)
}

// LessThan compares a.P/a.Q < b.P/b.Q using exact 128-bit cross-products (zero overflow).
func (a Rational) LessThan(b Rational) (bool, error) {
	if a.Q == 0 || b.Q == 0 {
		return false, ErrZeroDenominator
	}
	hi1, lo1 := bits.Mul64(a.P, b.Q)
	hi2, lo2 := bits.Mul64(b.P, a.Q)
	if hi1 != hi2 {
		return hi1 < hi2, nil
	}
	return lo1 < lo2, nil
}

func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
