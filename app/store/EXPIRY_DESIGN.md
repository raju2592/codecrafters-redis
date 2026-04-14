# Key Expiry Design

## Overview

Keys can be set with an optional expiry duration (e.g., `SET foo bar PX 100`). Expired keys are removed through two mechanisms: passive expiration on read, and active expiration via a background worker.

## Data Structures

### Global `sync.Map` — key-value store

Stores all key-value pairs as `key -> entry{value, expiresAt}`.

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

### `Set(key, value)` — no expiry

1. Write `entry{value, zero expiresAt}` to `sync.Map`.
2. Hash key to find shard. Lock shard.
3. If key exists in shard's map, swap-delete it from the slice and remove from map.
4. Unlock shard.

**Rationale:** Must remove from expiry tracking in case this key previously had an expiry.

### `SetWithExpiry(key, value, duration)`

1. Write `entry{value, time.Now().Add(duration)}` to `sync.Map`.
2. Hash key to find shard. Lock shard.
3. If key is not in shard's map, append to slice and record index in map.
4. Unlock shard.

**Rationale:** The map check (step 3) prevents duplicate entries in the slice when the same key is set with expiry multiple times. Without this, repeated `SetWithExpiry` calls would bloat the slice.

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

A single background goroutine runs on a ticker (e.g., every 100ms). On each tick:

1. Pick a random shard. Lock it.
2. Sample up to 20 random keys from the shard's slice.
3. For each sampled key, load from `sync.Map` and check expiry.
4. If expired: delete from `sync.Map`, swap-delete from shard's slice, remove from shard's map.
5. If not expired or not found: skip (key was overwritten or already deleted).
6. Unlock shard.

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
