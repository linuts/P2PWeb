# p2p Overlay Web Specification v0.1 (Draft)

**Status:** Draft

**Date:** 2025‑08‑12

**Editors:** (TBD)

**This document defines the p2p overlay web: a decentralized naming, distribution, and verification system for websites under the special-use suffix `.p2p`.**

The specification uses the key words **MUST**, **MUST NOT**, **REQUIRED**, **SHALL**, **SHALL NOT**, **SHOULD**, **SHOULD NOT**, **RECOMMENDED**, **MAY** per \[RFC 2119].

---

## 1. Scope and Goals

p2p provides an **overlay web** independent of ICANN DNS and public certificate authorities. It enables:

* Human‑memorable names under `.p2p`.
* Content‑addressed, verifiable site bundles.
* Multiple independently published variants of the same name, with a chooser UI and deterministic **hash clusters**.
* Peer‑to‑peer replication and discovery.

Non‑goals:

* Transparent impersonation of existing public domains.
* Centralized naming authorities.

---

## 2. Terminology

* **Peer:** A node speaking the p2p protocols.
* **Resolver:** Local service that intercepts `*.p2p` queries and performs name resolution.
* **Bundle:** A content‑addressed archive of a site’s files.
* **Manifest:** Signed metadata describing a bundle’s contents.
* **ManifestID:** SHA‑256 of the canonical CBOR bytes of the Manifest with signature field omitted.
* **Name Record (NR):** Signed mapping from `FQDN.p2p` to a ManifestID and publisher key.
* **Cluster:** Set of Name Records for the same FQDN that reference the same ManifestID.
* **Attestation:** Optional signed reputation statement about an `(FQDN, ManifestID)` pair.

---

## 3. Architecture Overview

p2p comprises four layers:

1. **Naming & Discovery:** A Kademlia DHT namespace for `*.p2p` Name Records. Optionally, libp2p PubSub for gossip.
2. **Content Distribution:** Retrieval of Bundles by ManifestID from peers. HTTP(S) gateway MAY be provided as a bootstrap aid.
3. **Verification:** Ed25519 signatures on Manifests and Name Records; SHA‑256 content digests for files; canonical CBOR for hash stability.
4. **Presentation:** Local HTTP(S) server provides same‑origin isolation at `https://<name>.p2p.localhost/` and `https://h-<shortid>.p2p.localhost/`. A **Chooser UI** presents clusters when multiple variants exist.

---

## 4. Special‑Use Name `.p2p`

### 4.1 Resolution Policy

* Hostnames ending in `.p2p` are **special‑use** and MUST be resolved exclusively by a local p2p Resolver.
* System configurations SHOULD be set to route `.p2p` DNS queries to `127.0.0.1:<port>` or platform‑specific equivalents (split DNS). Mobile implementations MAY use a VPN/TUN service to intercept requests.
* Resolvers MUST NOT forward `.p2p` to public recursive resolvers.

### 4.2 URL Forms

* Standard form: `https://example.p2p/` (presented to the user).
* Local origin form: `https://example.p2p.localhost/` (actual binding served by the Resolver’s local HTTP server).
* Manifest‑pinned origin: `https://h-<short_manifest_id>.p2p.localhost/`.

---

## 5. Cryptography

* **Public Key Algorithm:** Ed25519 (EdDSA over Curve25519).
* **Content Hash:** SHA‑256.
* **Encoding:** Canonical CBOR (DAG‑CBOR constraints: map keys deterministically ordered; avoid NaN variants; prohibit indefinite length encodings). Implementations MUST ensure canonical serialization to obtain stable hashes.
* **Key Representation:** 32‑byte Ed25519 public keys; binary in CBOR; base32/base58 for UI display.
* **Signatures:** 64‑byte Ed25519 signatures over canonical CBOR encodings of the unsigned structure.

---

## 6. Data Models

### 6.1 CDDL Definitions

```
; Manifest (unsigned fields only when computing ManifestID)
Manifest = {
  1: tstr,                 ; name ("example.p2p")
  2: uint,                 ; version
  3: { * tstr => bstr },   ; files: path -> sha256(file)
  4: tstr,                 ; entry path ("/index.html")
  5: bstr,                 ; publisher pubkey (32 bytes)
  6: uint,                 ; created_at (unix seconds)
  7: [ * Feed ],           ; optional feeds
  8: ? bstr                ; signature (64 bytes) -- OMITTED when hashing
}

Feed = {
  1: tstr,                 ; id ("comments")
  2: tstr,                 ; type (e.g., "crdt-automerge")
  3: bstr                  ; root content id / multihash
}

NameRecord = {
  1: tstr,                 ; fqdn ("example.p2p")
  2: bstr,                 ; manifest_id (sha256 of unsigned Manifest)
  3: bstr,                 ; publisher pubkey (32 bytes)
  4: uint,                 ; timestamp (unix seconds)
  5: [ * Peer ],           ; peers/mirrors (optional)
  6: ? bstr                ; signature (64 bytes)
}

Peer = {
  1: tstr,                 ; peer id (libp2p multihash)
  2: tstr                  ; multiaddr
}

Attest = {
  1: bstr,                 ; attester pubkey
  2: tstr,                 ; fqdn
  3: bstr,                 ; manifest_id
  4: nint,                 ; weight [-5..+5]
  5: uint,                 ; timestamp
  6: ? bstr                ; signature
}
```

### 6.2 Identifiers

* **ManifestID**: `sha256(cbor(Manifest_without_signature))`.
* **NameRecordID**: `sha256(cbor(NameRecord_without_signature))`.

---

## 7. Bundle Format

### 7.1 Requirements

* A Bundle MUST contain the full site files and a `manifest.cbor` (signed Manifest).
* Bundle containers MAY be TAR, ZIP, or CAR (Content Addressable aRchive). Implementations MUST record per‑file SHA‑256 digests in the Manifest.
* The entry point file **MUST** exist and be referenced by Manifest key 4.

### 7.2 Determinism

* Publishers SHOULD produce reproducible Bundles (stable file ordering, normalized metadata timestamps) so that independent builds yield identical Manifests and, therefore, identical ManifestIDs.

---

## 8. Publishing

### 8.1 Key Generation

* Publishers MUST generate Ed25519 keypairs. Private keys SHOULD be encrypted at rest.

### 8.2 Pack & Sign

* Compute SHA‑256 for each file; construct unsigned Manifest; produce canonical CBOR bytes; compute ManifestID; sign bytes with publisher private key and embed signature in field 8.

### 8.3 Announcing Name Records

* A publisher MUST create a Name Record mapping `FQDN.p2p` → `ManifestID` with its public key and current timestamp, sign it, and publish to the DHT under key `name:<FQDN>`.
* Multiple Name Records MAY exist for the same FQDN (e.g., from different publishers). Resolvers MUST be prepared to retrieve and aggregate all NRs present.

---

## 9. Discovery & Transport

### 9.1 DHT Namespace

* Implementations SHOULD use a libp2p Kademlia DHT. The DHT key for a name is the canonical byte string `name:<FQDN>`.
* The DHT value type is a CBOR array of one or more signed Name Records. Peers MAY chunk large sets.

### 9.2 Providers

* Peers that pin a Bundle SHOULD announce provider records for the associated ManifestID to enable efficient retrieval.

### 9.3 Retrieval

* Given a selected ManifestID, clients MUST retrieve `manifest.cbor` and referenced files from one or more providers, verifying each file’s digest against the Manifest.
* Implementations MAY support HTTP(S) gateways for bootstrap; cryptographic verification remains mandatory.

---

## 10. Resolution Algorithm

Given `FQDN.p2p`:

1. **Lookup NRs:** Query DHT key `name:<FQDN>`; parse and verify signatures. Discard invalid or stale records (see §12).
2. **Cluster:** Group valid NRs by `ManifestID`.
3. **Rank:** Compute cluster scores (see §11) and order descending.
4. **Choose:** If one cluster passes acceptance thresholds, clients MAY auto‑select; otherwise invoke the **Chooser UI** to let the user pick.
5. **Fetch:** Retrieve the Bundle for the chosen ManifestID, verify, and serve locally at `https://<FQDN>.p2p.localhost/`.

---

## 11. Ranking and Selection

### 11.1 Cluster Score

Implementations SHOULD compute a cluster score using the following components:

```
score = w1*ln(1 + distinct_publishers)
      + w2*ln(1 + distinct_peers)
      + w3*usage_weight
      + w4*freshness_decay
      + w5*attestation_sum
```

* **distinct\_publishers:** Count of unique publisher keys referencing the ManifestID for FQDN.
* **distinct\_peers:** Count of unique providers observed for the ManifestID.
* **usage\_weight:** Locally aggregated successful fetches/visits with privacy‑preserving telemetry (opt‑in).
* **freshness\_decay:** A decreasing function of NR timestamps; e.g., `exp(-Δt/τ)`.
* **attestation\_sum:** Sum of trusted attesters’ weights for `(FQDN, ManifestID)`.

### 11.2 Acceptance Thresholds

* Implementations MAY auto‑select the top cluster if `(score_top - score_next) ≥ δ` and `distinct_publishers ≥ P_min`.
* Otherwise, present the Chooser UI with per‑cluster metadata and **Select** actions.

---

## 12. Validity, Staleness, and Revocation

* A Name Record is **valid** if the signature verifies and `timestamp ≤ now + 5 min`.
* A Name Record is **fresh** if `now - timestamp ≤ T_NR` (RECOMMENDED default: 30 days). Stale NRs MAY be ignored or down‑weighted.
* Publishers MAY rotate keys via a **KeyLink** record signed by the old key mapping `old_pubkey → new_pubkey`. Resolvers SHOULD follow the most recent valid link when aggregating publisher identity statistics.
* Hard revocation is out‑of‑band: users MAY subscribe to blocklists or trust oracles. Implementations SHOULD honor configured blocklists during ranking and display.

---

## 13. Local Presentation & Origin Model

* The Resolver MUST serve chosen Bundles from a loopback HTTP(S) server.
* Each FQDN MUST map to origin `https://<FQDN>.p2p.localhost/`.
* Each ManifestID MUST also map to `https://h-<shortid>.p2p.localhost/` for strict isolation between variants.
* TLS MAY be self‑issued for `*.p2p.localhost`. User agents SHOULD pre‑trust the local certificate or use an application shell with an embedded webview.

---

## 14. Security Considerations

* **No CA Impersonation:** Implementations MUST NOT present themselves as authoritative for real public domains or obtain certificates for them.
* **Signature Verification:** Clients MUST verify Ed25519 signatures on Manifests and Name Records prior to use.
* **Content Integrity:** Clients MUST verify per‑file SHA‑256 digests before serving to the browser.
* **Sybil Resistance:** Ranking SHOULD incorporate publisher and peer diversity; optional Web‑of‑Trust Attestations mitigate sybils.
* **Reproducible Builds:** Publishers SHOULD use deterministic build pipelines to enable identical ManifestIDs across independent builds.
* **Sandboxing:** Serving via distinct manifest‑pinned origins prevents active content from one variant accessing another.

---

## 15. Privacy Considerations

* Telemetry for usage weighting MUST be opt‑in and aggregated locally or via privacy‑preserving methods. Raw browsing histories MUST NOT be broadcast.
* DHT queries MAY leak interest in a name. Implementations SHOULD support query batching and opportunistic caching.

---

## 16. Error Handling

* **NR\_INVALID\_SIGNATURE**: Discard record; MAY display warning in diagnostics.
* **NR\_STALE**: Down‑weight or ignore.
* **BUNDLE\_MISSING\_FILE**: Abort fetch; display error.
* **DIGEST\_MISMATCH**: Abort; mark provider as faulty.
* **MANIFEST\_SIG\_INVALID**: Reject bundle; do not present in UI.
* **NO\_PROVIDERS**: Offer retry/backoff and (if configured) gateway fallback.

---

## 17. Interoperability and Versioning

* This spec is versioned by **SpecVersion** = `0.1`. Future incompatible changes MUST bump the major version and alter CDDL keys or container media types to avoid ambiguity.
* Manifests MAY include an optional `spec_version` tstr key (future extension) guarded by a new CDDL field number.

---

## 18. Reference Operations (Non‑Normative)

### 18.1 Publisher CLI

* `p2p keygen` → Ed25519 keypair
* `p2p pack` → Bundle + signed Manifest, prints ManifestID
* `p2p publish` → DHT NR announce, provider add

### 18.2 Resolver

* Split‑DNS for `.p2p`
* DHT get → cluster → rank → choose → fetch → verify → serve

---

## 19. Security Profiles (MUST/SHOULD Sets)

**Strict Mode (RECOMMENDED):**

* Require `distinct_publishers ≥ 2` to auto‑select.
* Disable gateway fallback.
* Enforce reproducible‑build hint (reject Manifests without deterministic metadata flag).

**Permissive Mode:**

* Allow single‑publisher clusters; enable gateway fallback with explicit UI notice.

---

## 20. IANA Considerations (Informative)

`.p2p` is intended as a special‑use name resolved exclusively by local p2p Resolvers. Deployments MUST avoid leaking `.p2p` queries to public DNS.

---

## 21. Implementation Notes (Informative)

* libp2p Kademlia and PubSub are suitable substrates, but any DHT/transport satisfying §9 semantics MAY be used.
* CAR (Content Addressable aRchive) integrates well with content‑addressed stores, but ZIP/TAR are acceptable with strict hashing (§7).

---

## 22. Open Issues

1. Standardize Attestation trust oracles and distribution.
2. Define normative freshness parameters (`T_NR`, τ, δ, `P_min`, weights `w1..w5`).
3. Key rotation discovery and caching semantics.
4. Internationalized domain labels under `.p2p`.

---

## 23. References

* \[RFC 2119] Key words for use in RFCs to Indicate Requirement Levels.
* Ed25519: Bernstein et al., "High-speed high-security signatures".
* SHA‑256: FIPS 180‑4.
* CBOR: RFC 8949. DAG‑CBOR: IPFS/IPLD canonical CBOR guidelines.

---

## Appendix A: Example Encodings (Informative)

### A.1 Minimal Manifest (JSON notation for readability)

```
{
  "1": "docs.p2p",
  "2": 1,
  "3": {
    "/index.html": h'9F1B…',
    "/app.js":     h'12A7…'
  },
  "4": "/index.html",
  "5": h'F1E2…32bytes…',
  "6": 1723440000,
  "7": [],
  "8": h'64bytesig…'
}
```

### A.2 Name Record

```
{
  "1": "docs.p2p",
  "2": h'32byteManifestID',
  "3": h'publisherPubkey32',
  "4": 1723440100,
  "5": [ {"1": "12D3KooW…", "2": "/ip4/1.2.3.4/tcp/4001/p2p/12D3…"} ],
  "6": h'64bytesig…'
}
```

---

*End of Specification*

---

## Local Demo

This repository ships with a minimal DNS resolver and web server for experimenting with the `.p2p` special-use domain.

1. Start the combined demo server (DNS resolver and web server):
   ```sh
   go run cmd/p2pweb/main.go
   ```
2. Configure your operating system to use `127.1.1.153` as its DNS server (replace existing nameserver entries).
3. Visit `http://example.p2p:8080` in a browser or:
   ```sh
   curl http://example.p2p:8080
   ```
