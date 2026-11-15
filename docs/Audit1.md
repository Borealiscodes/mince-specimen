# NDH-META-SYSTEMS Architectural Analysis: Upstream Public Go Facade (`2836cb9`) vs. `mince-specimen`

**Target Specimen:** `Borealiscodes/mince-specimen` (Concurrent Fork of `8b-is/mince`) [1]  
**Upstream Reference Commit:** [`2836cb9facbef76c93ee931bfe72ada0d8b4d9bb`](https://github.com) (`feat: add public Go library facade (#17)`) [1]  
**Methodology & References:** This comparative analysis evaluates the architectural differences, mathematical guardrails, and cryptographic provenance models between the upstream repository and the `mince-specimen` fork under the NDH-Meta-Systems methodology (Zenodo DOI: `10.5281/zenodo.24149283`) [1].

---

## 1. Executive Synthesis

Upstream commit `2836cb9` (`8b-is/mince`, PR #17) transitions `mince` into an importable Go library (`module github.com/8b-is/mince`), exposing core entry points like `Mince` and `Store` while keeping implementation details inside Go's `internal/` boundary [1]. 

While upstream focuses on ergonomic local context folding, `Borealiscodes/mince-specimen` hardens the engine against three critical edge failure modes—silent disk history tampering, 64-bit integer wraparound, and unbounded context window retrieval—using zero-dependency modules in `refactor/internal/` [1].

---

## 2. Core Architectural Comparison

| Dimension | Upstream `8b-is/mince` (`2836cb9`) | `Borealiscodes/mince-specimen` |
| :--- | :--- | :--- |
| **Primary Role** | Pragmatic local CLI, HTTP/MCP proxy, and pre-v1 Go library facade (`github.com/8b-is/mince`) for deterministic context compression [1]. | Forensic systems specimen and zero-dependency cryptographic and mathematical hardening testbed [1]. |
| **Provenance & Tamper Evidence** | Uses a local SHA-256 hash chain and CRC32C (`internal/litany`), which detects accidental disk corruption but can be locally rewritten [1]. | Adds **Ed25519 asymmetric epoch checkpoints** (`refactor/internal/litany/ed25519_anchor.go`), binding chain heads and timestamps to a non-repudiable public key [1]. |
| **Rational Arithmetic ($p/q$)** | Uses standard 64-bit reduced fractions to avoid float drift, which can silently wrap modulo $2^{64}$ during long sessions [1]. | Implements **128-bit overflow-checked rationals** (`refactor/internal/felix/rational_guard.go`) via `math/bits`, failing closed (`ErrRationalOverflow`) on bounds violations [1]. |
| **Retrieval Safety (`Retrieve`)** | Returns requested slices without a hard protocol-level byte cap, risking large context dumps on broad queries [1]. | Enforces a hard **32 KiB byte ceiling** (`refactor/internal/wireagent/retrieve_cap.go`) via `io.LimitReader` with safety checks for valid UTF-8 boundaries [1]. |
| **Governance & Lineage** | Pragmatic MIT utility with a pre-v1 public Go contract [1]. | Anchored to the NDH-Meta-Systems Five-Boundary methodology and Zenodo DOI (`10.5281/zenodo.24149283`) [1]. |

---

## 3. Formal Invariants & Integration Mapping

* **Twisted-Edwards Chain Checkpointing (`refactor/internal/litany/ed25519_anchor.go`):** Upgrades local hashing to non-repudiable cryptographic authorship while preserving `CGO_ENABLED=0` zero-dependency builds [1].
* **Overflow-Proof 128-Bit Rationals (`refactor/internal/felix/rational_guard.go`):** Evaluates cross-multiplications in exact 128-bit unsigned arithmetic prior to Euclidean reduction to prevent silent modulo $2^{64}$ wrapping [1].
* **Bounded UTF-8 Retrieval (`refactor/internal/wireagent/retrieve_cap.go`):** Guarantees tool-retrieval outputs respect a $32{,}768\text{-byte}$ ceiling ($B_{\max}$) while preserving valid UTF-8 codepoint boundaries [1].
* **Facade Integration:** Because commit `2836cb9` exposes `Store.Retrieve` directly at the package root, wrapping it with `retrieve_cap.go` and utilizing `rational_guard.go` / `ed25519_anchor.go` provides clean, drop-in safety primitives without violating upstream API contracts [1].

---

## 4. Non-Activation & Provenance Clause

> All comparative analyses, architectural mappings, and formal invariants documented in this file are descriptive and forensic-only. They do not activate NDH geometry, holonomy engines, constellation routing, lineage engines, or governance altitude [1].
