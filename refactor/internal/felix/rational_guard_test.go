// Package felix tests overflow-checked exact p/q rational arithmetic
// against unguarded 64-bit wraparound and math/big.Rat ground truth.
// Methodology: NDH-Meta-Systems (Sealed-Layer Neutral)
package felix

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

// legacyAddUint64 reproduces standard 64-bit rational addition that silently wraps mod 2^64.
func legacyAddUint64(a, b Rational) Rational {
	num := a.P*b.Q + b.P*a.Q
	den := a.Q * b.Q
	if den == 0 {
		return Rational{P: 0, Q: 0}
	}
	g := gcd(num, den)
	return Rational{P: num / g, Q: den / g}
}

func TestNewRational_ConstructionAndZeroGuards(t *testing.T) {
	if _, err := NewRational(1, 0); !errors.Is(err, ErrZeroDenominator) {
		t.Fatalf("expected ErrZeroDenominator, got %v", err)
	}

	zero, err := NewRational(0, 42)
	if err != nil {
		t.Fatalf("unexpected error for zero numerator: %v", err)
	}
	if zero.P != 0 || zero.Q != 1 {
		t.Fatalf("expected canonical 0/1, got %d/%d", zero.P, zero.Q)
	}

	reduced, err := NewRational(18, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reduced.P != 3 || reduced.Q != 4 {
		t.Fatalf("expected reduced 3/4, got %d/%d", reduced.P, reduced.Q)
	}
}

func TestRationalAdd_PreventsSilentWraparound(t *testing.T) {
	// Two valid uint64 fractions with large coprime denominators whose product exceeds 2^64 - 1.
	a, err := NewRational(3, (1<<33)-9)
	if err != nil {
		t.Fatalf("setup a failed: %v", err)
	}
	b, err := NewRational(5, (1<<33)-17)
	if err != nil {
		t.Fatalf("setup b failed: %v", err)
	}

	// Ground truth via math/big.Rat
	exactA := new(big.Rat).SetFrac(new(big.Int).SetUint64(a.P), new(big.Int).SetUint64(a.Q))
	exactB := new(big.Rat).SetFrac(new(big.Int).SetUint64(b.P), new(big.Int).SetUint64(b.Q))
	exactSum := new(big.Rat).Add(exactA, exactB)

	// 1. Legacy 64-bit arithmetic silently wraps without returning an error
	legacy := legacyAddUint64(a, b)
	if exactSum.Num().IsUint64() && legacy.P == exactSum.Num().Uint64() && legacy.Q == exactSum.Denom().Uint64() {
		t.Fatalf("expected legacy uint64 addition to silently wrap, got exact %d/%d", legacy.P, legacy.Q)
	}
	t.Logf("Legacy uint64 silently wrapped: %d/%d (exact: %s)", legacy.P, legacy.Q, exactSum.RatString())

	// 2. Guarded Add fails closed with ErrRationalOverflow
	got, err := a.Add(b)
	if !errors.Is(err, ErrRationalOverflow) {
		t.Fatalf("expected ErrRationalOverflow, got %+v (err=%v)", got, err)
	}
}

func TestRationalMul_CrossPreReductionAvoidsFalseOverflow(t *testing.T) {
	// Direct multiplication (a.P * b.P) and (a.Q * b.Q) would overflow uint64,
	// but cross-GCD pre-reduction in Mul succeeds cleanly.
	const bigFactor = uint64(1) << 40
	a, err := NewRational(bigFactor, 3)
	if err != nil {
		t.Fatalf("setup a failed: %v", err)
	}
	b, err := NewRational(3, bigFactor)
	if err != nil {
		t.Fatalf("setup b failed: %v", err)
	}

	got, err := a.Mul(b)
	if err != nil {
		t.Fatalf("expected cross-reduced Mul to succeed, got err=%v", err)
	}
	if got.P != 1 || got.Q != 1 {
		t.Fatalf("expected 1/1, got %d/%d", got.P, got.Q)
	}

	// Irreducible product exceeding uint64 still fails closed
	maxRat := Rational{P: math.MaxUint64, Q: 1}
	two := Rational{P: 2, Q: 1}
	if _, err := maxRat.Mul(two); !errors.Is(err, ErrRationalOverflow) {
		t.Fatalf("expected ErrRationalOverflow on MaxUint64 * 2, got %v", err)
	}
}

func TestRationalLessThan_Full128BitComparisonWithoutOverflow(t *testing.T) {
	// Cross-products a.P*b.Q and b.P*a.Q both exceed 2^64 - 1, yet LessThan compares exact 128-bit values.
	a := Rational{P: math.MaxUint64 - 2, Q: math.MaxUint64 - 1}
	b := Rational{P: math.MaxUint64 - 1, Q: math.MaxUint64}

	less, err := a.LessThan(b)
	if err != nil {
		t.Fatalf("unexpected error in 128-bit LessThan: %v", err)
	}
	if !less {
		t.Fatalf("expected (%d/%d) < (%d/%d) to be true", a.P, a.Q, b.P, b.Q)
	}

	greaterOrEq, err := b.LessThan(a)
	if err != nil {
		t.Fatalf("unexpected error in reverse LessThan: %v", err)
	}
	if greaterOrEq {
		t.Fatalf("expected (%d/%d) < (%d/%d) to be false", b.P, b.Q, a.P, a.Q)
	}
}
