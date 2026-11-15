# NDH-META-SYSTEMS Final Synthesis: Architectural Audit, Remediation Matrix & Forward Risk Horizon

**Target Specimen:** `Borealiscodes/mince-specimen` (Concurrent Peer-Node Fork of `8b-is/mince`)  
**Upstream Baseline Commits:** `b4e4ba0` • `dda612b` • `2836cb9` (`://github.com`)  
**Methodology:** NDH-Meta-Systems (Five-Boundary Forensic Audit & Concurrent Peer-Node Hardening)  
**Canonical Anchor:** [DOI: 10.5281/zenodo.24149283](https://doi.org)  
**Status:** Sealed-Layer Neutral • Descriptive & Forensic-Only  

---

## 1. Executive Synthesis

Upstream `8b-is/mince` delivers a fast, single-binary (`CGO_ENABLED=0`, zero-dependency) local proxy and public Go facade (`commit 2836cb9`) for deterministic context compression and byte-exact KV-cache preservation. However, forensic inspection of upstream's storage, arithmetic, retrieval, and provenance layers reveals seven structural edge defects that degrade reliability, safety, and auditability at scale.

`Borealiscodes/mince-specimen` resolves these seven defects through **five zero-dependency Go hardening modules** and companion unit test suites under `refactor/internal/`, maintaining full compatibility with upstream's standard-library-only philosophy.

---

## 2. Complete Audit & Remediation Matrix

| Finding ID | Upstream Component | Structural Defect / Edge Failure Mode | Specimen Module (`refactor/internal/`) | Hardening Invariant & Test Suite |
| :--- | :--- | :--- | :--- | :--- |
| **LTNY-01** | `internal/litany` | **Unauthenticated Hash Chain:** Local SHA-256 + CRC32C chain detects accidental bit-rot, but any local process can rewrite or truncate history without detection. | `litany/ed25519_anchor.go` | **Ed25519 Epoch Checkpoints:** Signs canonical 84-byte frames (`LTNY \| epoch \| ts \| head \| signerID`) bound to `SHA256(PubKey)`. Verified via `ed25519_anchor_test.go`. |
| **FELX-01** | `internal/felix` & `internal/pq` | **Silent 64-Bit Rational Wraparound:** Standard `uint64` cross-multiplication (`p1*q2 + p2*q1`) silently wraps modulo $2^{64}$ on long-running sessions. | `felix/rational_guard.go` | **128-Bit Checked Rationals:** Evaluates cross-products via `math/bits.Mul64` / `Add64`, pre-reduces via GCD, and fails closed (`ErrRationalOverflow`). Verified via `rational_guard_test.go`. |
| **WIRE-01** | `internal/wireagent` & `ccr.Retrieve` | **Unbounded Context Dump & UTF-8 Slicing:** Broad queries return megabytes into LLM context; raw byte cuts corrupt multi-byte UTF-8 runes. | `wireagent/retrieve_cap.go` | **32 KiB Ceiling + UTF-8 Alignment:** Clamps reads via `io.LimitReader` (`<= 128 KiB` hard cap) and steps back `<= 3` bytes for `utf8.Valid`. Verified via `retrieve_cap_test.go`. |
| **CCR-01** | `internal/storage/ccr/store.go` (`Put`) | **Eager Pre-Accept Disk Write Leak:** Writes `ccr_<id>.gz` to disk *before* verifying 10% savings, stranding orphaned files with no `Delete()` API. | `storage/commit_guard.go` | **Two-Phase Commit-on-Accept:** Stages up to `16 MiB` in RAM (`Stage`), writes to disk *only* on net reduction (`CommitIfAccepted`), and zero-scrubs memory on reject. Verified via `commit_guard_test.go`. |
| **CCR-02** | `internal/storage/ccr/store.go` (`path`) | **Unsanitized `id` Path Traversal:** Concatenates raw `id` directly into `filepath.Join(s.dir, "ccr_"+id+".gz")` without validation. | `storage/hardened_ccr.go` | **Strict Hex-ID Normalization:** `NormalizeCCRID` enforces `12..64` lowercase `[0-9a-f]` characters, rejecting `../` and malformed IDs (`ErrInvalidCCRID`). Verified via `hardened_ccr_test.go`. |
| **CCR-03** | `internal/storage/ccr/store.go` (`writeDisk`) | **Static `.tmp` Race Condition:** Uses deterministic `s.path(id) + ".tmp"` without holding mutex across disk I/O, causing concurrent rename collisions. | `storage/hardened_ccr.go` | **Unique Temp-File Atomic Rename:** Uses `os.CreateTemp(dir, ".ccr-*.tmp")` (`0600`) with deferred cleanup before `os.Rename`. Verified under `-race` in `hardened_ccr_test.go`. |
| **CCR-04** | `internal/storage/ccr/store.go` (`Get`) | **Mutable LRU Slice Aliasing:** `Get(id)` returns the internal `memEntry.data` slice header directly without defensive copying. | `storage/hardened_ccr.go` & `commit_guard.go` | **Defensive Buffer Isolation:** Allocates fresh byte slices across staging and verified reads so caller mutation cannot corrupt cached records. |

---

## 3. Forward Risk Horizon (Watchlist for Future Upstream Commits)

As `8b-is/mince` evolves past its pre-v1 public Go facade (`commit 2836cb9`), five unresolved architectural risks remain on the watchlist for future specimen audits:

1. **Unbounded Disk Growth & Missing Retention/GC (`CCR-05`):** Even with `commit_guard.go` preventing rejected-candidate leaks, accepted `ccr_<id>.gz` files currently have no TTL, LRU disk quota, or compaction sweep—posing gradual exhaustion risks on edge storage workstations.
2. **Decompression Bomb Exposure in `readDisk` (`CCR-06`):** Upstream's `io.ReadAll(zr)` reads uncompressed gzip streams into RAM without an `io.LimitReader` ceiling, allowing a crafted or corrupted `.gz` archive with a high compression ratio to trigger OOM.
3. **Unbounded Regex Compilation in `Retrieve` (`WIRE-02`):** `regexp.Compile(query)` accepts arbitrary caller/agent regex strings without length caps or complexity bounds, exposing long-running proxy processes to CPU stalls on pathological patterns.
4. **Disconnected Policy & Budget Enforcement (`CFG-01`):** As documented in upstream `commit dda612b`, `mince.vkd.hcl` capability budgets, salience weights, and `ralph` improvement gates are parsed at startup but remain unwired from live proxy routing—creating potential governance drift between declared policy and runtime execution.
5. **Pre-v1 Public API Churn (`API-01`):** Upstream explicitly marks `://github.com` as pre-v1 without stability guarantees; keeping `refactor/internal/` decoupled behind narrow interfaces (`BackingStore`, `io.Reader`) insulates `mince-specimen` from breaking upstream signature changes.

---

## 4. Non-Activation & Governance Clause

> All audit findings, remediation modules, test suites, and forward risk assessments in this synthesis are descriptive and forensic-only. They do not activate NDH geometry, holonomy engines, constellation routing, lineage engines, or governance altitude.
