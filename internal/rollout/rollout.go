package rollout

import "hash/fnv"

// Evaluate deterministically decides whether the rollout applies to the given
// user. The decision is stable: the same (key, user) pair always yields the
// same result. No randomness or time-based seed is involved.
//
//   - enabled == false            -> always false
//   - rolloutPercent == 100       -> always true (when enabled)
//   - rolloutPercent == 0         -> always false
//   - otherwise                   -> FNV-64a(key + "\x00" + user) % 100 < rolloutPercent
func Evaluate(key, user string, rolloutPercent int, enabled bool) bool {
	if !enabled {
		return false
	}
	if rolloutPercent >= 100 {
		return true
	}
	if rolloutPercent <= 0 {
		return false
	}

	h := fnv.New64a()
	h.Write([]byte(key))
	h.Write([]byte{0})
	h.Write([]byte(user))

	return int(h.Sum64()%100) < rolloutPercent
}
