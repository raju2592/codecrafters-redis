# Key Expiry Design

## Overview

Keys can be set with an optional expiry duration (e.g., `SET foo bar PX 100`). Expired keys are removed through two mechanisms: passive expiration on read, and active expiration via a background worker.

## Data Structures

### Global `sync.Map` — key-value store

Stores all key-value pairs as `key -> valueWithMeta{value, expiresAt}`.

**Rationale:** `sync.Map` provides lock-free reads in the common case (non-expired key lookups), which is the dominant operation in a Redis-like workload. This avoids holding any mutex on the read hot path.

### Shard array — expiry tracking

An array of N shards (e.g., 64), each containing:

- `sync.Mutex` — protects the shard's map and slice
- `map[string]int` — maps expiring key to its index in the slice
- `[]string` — ordered list of expiring keys for random sampling

**Rationale:**

- **Why sharded?** A single global lock for expiry tracking would serialize all `Set` calls. Sharding by key hash allows concurrent writes to different shards.
- **Why a map+slice combo?** The slice enables O(1) random sampling (pick a random index). The map enables O(1) dedup (check if key is already tracked) and O(1) removal (swap-delete using the stored index). Neither data structure alone provides both.
- **Why not store expiry data here?** The expiry timestamp lives in the `sync.Map` entry, not in the shard. This avoids needing to lock a shard during `Get` — the `sync.Map` read is lock-free.

## Operations

### `Set(key, value, options)`

A single function handles both expiry and non-expiry cases based on `SetOptions.Ttl`.

1. Hash key to find shard. Lock shard.
2. Create `valueWithMeta{value, expiresAt}`. If `Ttl > 0`, set `expiresAt = time.Now() + Ttl`.
3. Write to `sync.Map`.
4. If `Ttl > 0`: add key to shard's slice+map if not already present (dedup via map check).
5. If `Ttl == 0`: remove key from shard's slice+map if present (in case it previously had an expiry).
6. Unlock shard.

**Rationale:** The map check (step 4) prevents duplicate entries in the slice when the same key is set with expiry multiple times. Step 5 cleans up expiry tracking when a key is overwritten without a TTL.

### `Get(key)` — passive expiration

1. Load entry from `sync.Map` (lock-free).
2. If not found, return miss.
3. If `expiresAt` is non-zero and in the past, return miss and send the key to the passive expiry channel (non-blocking).
4. Otherwise return the value.

**Rationale:** No locks are taken on the read path. The actual deletion is deferred to passive expiry consumers via a buffered channel. This keeps `Get` fast even when encountering expired keys. The channel send is non-blocking (`select` with `default`) so `Get` never stalls if the channel is full — the active expiry worker will eventually catch the key anyway.

### Passive expiry consumers

A pool of consumer goroutines (e.g., 3-5) reads from the passive expiry channel. Each consumer:

1. Receives a key from the channel.
2. Calls `deleteIfExpired(key)`: locks the shard, re-checks the entry in `sync.Map`, deletes if still expired, updates the shard's slice+map.

**Rationale:** Separating passive cleanup from `Get` keeps reads lock-free. Multiple consumers drain the channel faster, reducing the chance of a full buffer. The consumer count is kept small (3-5) to avoid excessive lock contention on the shards — with 64 shards, diminishing returns kick in quickly.

### Active expiry worker

A single background goroutine runs on a ticker (every 1000ms). On each tick:

1. Try up to `N_RANDOM_SAMPLE_SIZE * 2` iterations to collect `N_RANDOM_SAMPLE_SIZE` (20) keys:
   a. Pick a random shard. Lock it.
   b. If shard's slice is empty, unlock and try again.
   c. Pick a random key from the shard's slice. Record it. Unlock.
2. For each collected key, call `deleteIfExpired(key)` which locks the shard, re-checks the entry in `sync.Map`, and deletes if still expired (from both `sync.Map` and shard's slice+map).

**Rationale:** Passive expiration alone is insufficient — keys that are never read again would leak memory. The active worker ensures all expired keys are eventually cleaned up. Random sampling (rather than full scans) keeps the cost bounded per tick, matching Redis's approach. The active worker is fully independent from the passive consumers — it does not interact with the channel.

## Concurrency Analysis

| Operation | `sync.Map` | Shard lock | Channel |
|---|---|---|---|
| `Get` (non-expired) | Lock-free read | None | None |
| `Get` (expired) | Lock-free read | None | Non-blocking send |
| `Set` | Atomic store | Shard mutex | None |
| `SetWithExpiry` | Atomic store | Shard mutex | None |
| Passive consumer | Atomic load/delete | Shard mutex | Blocking receive |
| Active expiry worker | Atomic load/delete | Shard mutex | None |

The critical invariant: the `sync.Map` and shard state can be momentarily inconsistent (e.g., a key is in the shard's slice but already deleted from `sync.Map`). This is safe because:

- The shard is the authority on *which keys to check*, not *whether they're expired*.
- The `sync.Map` entry is the authority on *whether a key is expired*.
- Stale entries in the shard are harmlessly skipped when the `sync.Map` lookup misses.

### TOCTOU safety

The active expiry worker and queued deletes always lock the shard and re-check the entry in `sync.Map` before deleting. This prevents a race where:

1. Worker sees key is expired
2. Concurrent `SetWithExpiry` writes a new value for the same key
3. Worker deletes the new value

The shard lock serializes step 1 and step 2 for keys in the same shard. After acquiring the lock, the worker re-reads from `sync.Map` to confirm the entry is still expired.
