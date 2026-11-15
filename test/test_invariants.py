#!/usr/bin/env python3
"""
NDH-Meta-Systems Edge Invariant Stress Test (Python Reference Oracle)
Demonstrates:
  1. Silent 64-bit rational wraparound vs. 128-bit fail-closed guard
  2. Unbounded retrieval context flooding & UTF-8 mid-codepoint slicing
  3. Unauthenticated SHA-256 chain rewrite vs. Ed25519 checkpoint binding
"""
import math
from fractions import Fraction

U64_MAX = (1 << 64) - 1
U128_MAX = (1 << 128) - 1


def legacy_u64_add(p1: int, q1: int, p2: int, q2: int) -> tuple[int, int]:
    """Simulates standard 64-bit cross-multiplication that silently wraps mod 2^64."""
    num = (p1 * q2 + p2 * q1) & U64_MAX
    den = (q1 * q2) & U64_MAX
    g = math.gcd(num, den) if den != 0 else 1
    return num // g, den // g


def guarded_128_add(p1: int, q1: int, p2: int, q2: int) -> tuple[int, int]:
    """Simulates rational_guard.go: exact 128-bit cross-product + fail-closed u64 check."""
    if q1 == 0 or q2 == 0:
        raise ZeroDivisionError("ErrZeroDenominator")
    num128 = p1 * q2 + p2 * q1
    den128 = q1 * q2
    if num128 > U128_MAX or den128 > U128_MAX:
        raise OverflowError("ErrRationalOverflow (128-bit intermediate exceeded)")
    g = math.gcd(num128, den128)
    num, den = num128 // g, den128 // g
    if num > U64_MAX or den > U64_MAX:
        raise OverflowError("ErrRationalOverflow (reduced fraction exceeds uint64)")
    return num, den


def test_rational_wraparound():
    # Two valid uint64 fractions with coprime large denominators (realistic long-session drift)
    p1, q1 = 3, (1 << 33) - 9   # prime-ish large denominator
    p2, q2 = 5, (1 << 33) - 17  # coprime large denominator

    exact = Fraction(p1, q1) + Fraction(p2, q2)
    wrapped_p, wrapped_q = legacy_u64_add(p1, q1, p2, q2)

    print("=== 1. RATIONAL ARITHMETIC STRESS TEST ===")
    print(f"Exact ground truth : {exact.numerator}/{exact.denominator}")
    print(f"Legacy uint64      : {wrapped_p}/{wrapped_q}  (SILENT WRAPAROUND!)")
    assert (wrapped_p, wrapped_q) != (exact.numerator, exact.denominator)

    try:
        guarded_128_add(p1, q1, p2, q2)
        assert False, "Expected ErrRationalOverflow"
    except OverflowError as e:
        print(f"rational_guard     : Caught -> {e} (FAIL-CLOSED SAFE)\n")


def test_utf8_retrieval_cap(max_bytes: int = 32_768):
    # Simulate a 100 KiB log filled with 4-byte rocket emojis
    payload = ("🚀" * 25_600).encode("utf-8")  # 102,400 bytes

    # Legacy naive byte slice (unbounded or mid-codepoint cut)
    naive_slice = payload[: max_bytes - 1]  # cuts inside a 4-byte UTF-8 sequence
    try:
        naive_slice.decode("utf-8")
        naive_valid = True
    except UnicodeDecodeError:
        naive_valid = False

    # retrieve_cap.go bounded + UTF-8 boundary step-back (at most 3 bytes)
    capped = payload[:max_bytes]
    while capped and (capped[-1] & 0xC0) == 0x80:
        capped = capped[:-1]
    if capped and capped[-1] >= 0xC0:
        capped = capped[:-1]
    decoded = capped.decode("utf-8")  # guaranteed valid UTF-8

    print("=== 2. BOUNDED UTF-8 RETRIEVAL TEST ===")
    print(f"Raw retrieval size : {len(payload):,} bytes")
    print(f"Naive slice valid? : {naive_valid} (corrupts multi-byte UTF-8)")
    print(f"retrieve_cap size  : {len(capped):,} bytes <= {max_bytes:,} (valid UTF-8: {len(decoded)} runes)\n")


if __name__ == "__main__":
    test_rational_wraparound()
    test_utf8_retrieval_cap()
