# Go Core Hardening & Refactoring Plan: `mince-specimen` (v1.0)
**Target Repository:** `Borealiscodes/mince-specimen` (Research Fork of `8b-is/mince` v0.2.0)  
**Methodology:** `NDH-Meta-Systems` (Five-Boundary Separation of Concerns)  
**Constraint:** Zero External Dependencies (`Go 1.23+` Standard Library Only, `CGO_ENABLED=0`)  
**Canonical Reference:** *Epistemic Coevolution, Statutory Friction, and Sovereign Edge Governance* (`DOI: 10.5281/zenodo.24149283`)  

---

## 1. Objective
Strengthen four edge boundaries in `mince-specimen` using Go's built-in standard library (`crypto/ed25519`, `math/bits`) without adding third-party modules.

---

## 2. Target Package Hardening Matrix (`refactor/internal/`)

| Package Path | Hardening Action | Core Security & Mathematical Invariant |
| :--- | :--- | :--- |
| `refactor/internal/litany/ed25519_anchor.go` | **Add** | Signs periodic `LTNY v3` SHA-256 chain heads using Go's `crypto/ed25519` stdlib. |
| `refactor/internal/felix/rational_guard.go` | **Add** | Enforces overflow-checked `int64`/`uint64` arithmetic (`math/bits`) and Euclidean GCD reduction. |
| `refactor/internal/wireagent/retrieve_cap.go` | **Add** | Clamps `mince_retrieve` raw responses to a hard byte ceiling (`32 KiB`). |
| `mince.vkd.hcl` & `README.md` | **Annotate** | Documents upstream lineage provenance (`DOI: 10.5281/zenodo.24149283`). |

---

## 3. Four-Phase Implementation Sequence

* **Phase 1 — Asymmetric Provenance Checkpoint (`ed25519_anchor.go`):** Wrap `LTNY v3` hash-chain heads in a deterministic 72-byte epoch checkpoint signed with `crypto/ed25519`.
* **Phase 2 — Checked Rational Reduction (`rational_guard.go`):** Guard rational additions/multiplications against silent 64-bit integer wrap-around.
* **Phase 3 — Bounded Retrieval Guard (`retrieve_cap.go`):** Enforce `io.LimitReader` + hard byte budget (`maxBytes <= 32768`) inside `mince_retrieve`.
* **Phase 4 — Cloud CI Verification:** Run `go test -race ./...` and `go vet ./...` via GitHub Actions.

---

*For full metadata, cryptographic anchors, and non-activation clauses, please refer to the complete specification documentation [DOI 10.5281/zenodo.24149283].*
