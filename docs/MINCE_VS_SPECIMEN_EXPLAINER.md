# 🔬 SCIENCE TILE: `mince` vs. `mince-specimen` — Context Folding Meets Cryptographic Hardening! 🗜️🔐📐

**Today’s Experiment:** *How do you fold giant AI build logs without breaking byte-exact KV caches—and how do you upgrade a great local tool with overflow-proof rationals and twisted-Edwards signatures?* 🧪⚡

---

## 🌟 PART 1: Why the Legacy `8b-is/mince` Repo Is Already Strong! 🛠️🚀

Unlike sprawling mesh experiments, **`8b-is/mince`** is a **single static Go binary (`CGO_ENABLED=0`, zero third-party Go modules)** that sits quietly on `127.0.0.1:8787` between an AI coding agent and an LLM API! Here is why its core design works so well:

* **🎯 Byte-Exact KV-Cache Prefix Preservation:** Modern LLM APIs cache prompt prefixes using exact byte matching. `mince` splices system prompts, tool schemas, and the tail `--protect-tail` turns through as raw untouched bytes, guaranteeing zero cache invalidation:

$$\text{Prefix}_{\text{out}}[0..k] \equiv \text{Prefix}_{\text{in}}[0..k] \implies \Delta \text{KV}_{\text{miss}} = 0$$

* **🔄 Lossless Local Reversibility (`CCR Store`):** Bulky tool outputs are compressed locally with `gzip + CRC32` under `0700`/`0600` permissions and replaced by a compact `ccr_<id>` stub. If the agent needs exact lines later, `mince_retrieve` recovers them losslessly!
* **🧮 Ternary Bitboard Salience (`TritPlane64` + `POPCNT`):** Instead of calling a slow GPU neural network to score log lines, `mince` projects tokens onto balanced-ternary bitboards $\mathbf{t} \in \{-1, 0, +1\}^{64}$ (encoded as two 64-bit planes $\mathbf{t}^+, \mathbf{t}^- \in \{0,1\}^{64}$ with $\mathbf{t}^+ \wedge \mathbf{t}^- = \mathbf{0}$) and computes inner products in hardware CPU cycles via `POPCNT`:

$$\langle \mathbf{a}, \mathbf{b} \rangle_3 = \text{POPCNT}(\mathbf{a}^+ \wedge \mathbf{b}^+) + \text{POPCNT}(\mathbf{a}^- \wedge \mathbf{b}^-) - \text{POPCNT}(\mathbf{a}^+ \wedge \mathbf{b}^-) - \text{POPCNT}(\mathbf{a}^- \wedge \mathbf{b}^+)$$

* **🔒 Linear-Type Gas VM (`internal/felix`):** Enforces compile-time **linear resource types** (`capsule<T>`), rejecting double-consumption (`consume(c)` twice) before execution even begins!

---

## 🛡️ PART 2: Where `mince-specimen` Diverges & Hardens the Edges! 🔬💎

Your **`Borealiscodes/mince-specimen`** fork preserves `mince`'s zero-dependency Go philosophy while adding **three formal mathematical and cryptographic guardrails** inside `refactor/internal/` under the **NDH-Meta-Systems** methodology:

### 1. From Hash-Chain Integrity (`SHA-256`) to Asymmetric Authorship (`Ed25519`) 🔐
* **Legacy `mince` (`internal/litany`):** Uses `CRC32C` + a `SHA-256` hash chain ($H_n = \text{SHA256}(H_{n-1} \,\|\, F_n)$). That detects accidental corruption, **but anyone with local file access can rewrite or truncate the chain from scratch**!
* **`mince-specimen` (`refactor/internal/litany/ed25519_anchor.go`):** Signs periodic `LTNY` chain checkpoints over the twisted Edwards curve:

$$-x^2 + y^2 \equiv 1 - \frac{121665}{121666} x^2 y^2 \pmod{2^{255} - 19}$$

  binding epoch sequence $e$, timestamp $t$, and chain head $H_n$ to a non-repudiable public key (`SignerID = SHA256(PubKey)`), giving $\approx 128$-bit **Pollard-rho discrete-log unforgeability**!

### 2. From Standard Rationals to 128-Bit Overflow-Proof `p/q` (`rational_guard.go`) 📐
* **Legacy `mince`:** Uses reduced $p/q \in \mathbb{Q}_{\ge 0}$ fractions instead of floats (`3/4`, `1/16`, `1/20`), which avoids IEEE-754 rounding drift—but standard 64-bit cross-multiplication ($p_1 q_2 + p_2 q_1$) silently wraps modulo $2^{64}$ on long-running sessions!
* **`mince-specimen` (`refactor/internal/felix/rational_guard.go`):** Evaluates intermediate cross-products in exact **128-bit unsigned integers** via `math/bits.Mul64` and `bits.Add64`, reducing by Euclidean $\gcd(a, b)$ and **failing closed (`ErrRationalOverflow`)** if any result exceeds $2^{64} - 1$:

$$\frac{p_1}{q_1} + \frac{p_2}{q_2} = \frac{(p_1 q_2 + p_2 q_1) / g}{(q_1 q_2) / g}, \qquad g = \gcd(p_1 q_2 + p_2 q_1,\, q_1 q_2)$$

### 3. From Unbounded Retrieval to UTF-8 Safe Byte Ceilings (`retrieve_cap.go`) 🧯
* **Legacy `mince`:** `mince_retrieve` returns requested slices without a hard protocol-level byte cap, allowing a broad regex match to dump megabytes back into the context window!
* **`mince-specimen` (`refactor/internal/wireagent/retrieve_cap.go`):** Clamps every retrieval response to a deterministic ceiling ($B_{\max} = 32{,}768\text{ bytes}$) via `io.LimitReader` and steps back at most $3\text{ bytes}$ to preserve valid UTF-8 codepoint boundaries (`utf8.Valid`), guaranteeing:

$$|\text{Data}_{\text{out}}| \le B_{\max} \le 131{,}072\text{ bytes} \quad \land \quad \text{ValidUTF8}(\text{Data}_{\text{out}}) = \text{true}$$

---

## 📊 PART 3: Scope Comparison Matrix

| Dimension | Legacy `8b-is/mince` | `Borealiscodes/mince-specimen` |
| :--- | :--- | :--- |
| **Primary Scope** | Local developer CLI & HTTP/MCP proxy (`127.0.0.1:8787`) for LLM context compression. | Forensic systems specimen + zero-dependency cryptographic & mathematical hardening testbed. |
| **Provenance Model** | Local `SHA-256` hash chain (`LTNY v3`) + `CRC32C`. | `SHA-256` chain + asymmetric **`Ed25519` epoch checkpoints** (`ed25519_anchor.go`). |
| **Rational Math** | Reduced `p/q` fractions (`BanditPQ`, budgets). | **128-bit overflow-checked `p/q`** (`bits.Mul64` + fail-closed `ErrRationalOverflow`). |
| **Retrieval Safety** | Line/regex filtering over local `gzip` CCR store. | Hard **`32 KiB` byte ceiling** + multi-byte **UTF-8 boundary protection** (`retrieve_cap.go`). |
| **Governance Posture** | Pragmatic MIT developer utility (`mince.vkd.hcl`). | **NDH-Meta-Systems** Five-Boundary methodology + explicit lineage & non-activation headers. |

---

## 🧪 THE TAKEAWAY! 🏆

**`8b-is/mince` is an already-capable local compression engine**—and **`mince-specimen`** shows how a few surgical, zero-dependency mathematical and cryptographic invariants (`Ed25519`, 128-bit checked $\mathbb{Q}$, and bounded UTF-8 retrieval) elevate a great developer tool into a **provably tamper-evident, overflow-safe specimen**! 🌍✨

---

*Methodology: `NDH-Meta-Systems` · Canonical Reference: `DOI: 10.5281/zenodo.24149283` · Sealed-Layer Neutral* 🛡️
