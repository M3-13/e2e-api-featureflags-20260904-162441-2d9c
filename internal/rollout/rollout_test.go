package rollout

import (
	"fmt"
	"testing"
)

func TestEvaluateStability(t *testing.T) {
	// The same (key, user) pair must always produce the same result.
	inputs := []struct {
		key     string
		user    string
		percent int
	}{
		{"feature-x", "alice", 50},
		{"feature-x", "bob", 50},
		{"feature-y", "alice", 25},
		{"feature-z", "user-123", 75},
	}

	for _, in := range inputs {
		first := Evaluate(in.key, in.user, in.percent, true)
		for i := 0; i < 100; i++ {
			if got := Evaluate(in.key, in.user, in.percent, true); got != first {
				t.Fatalf("Evaluate(%q, %q, %d) not stable: got %v then %v", in.key, in.user, in.percent, first, got)
			}
		}
	}
}

func TestEvaluateDisabledAlwaysFalse(t *testing.T) {
	for _, percent := range []int{0, 1, 50, 99, 100} {
		if got := Evaluate("feature", "alice", percent, false); got {
			t.Errorf("enabled=false with percent=%d: got true, want false", percent)
		}
	}
}

func TestEvaluateFullRolloutAlwaysTrue(t *testing.T) {
	if got := Evaluate("feature", "alice", 100, true); !got {
		t.Errorf("enabled=true, percent=100: got false, want true")
	}
}

func TestEvaluateZeroPercentAlwaysFalse(t *testing.T) {
	if got := Evaluate("feature", "alice", 0, true); got {
		t.Errorf("enabled=true, percent=0: got true, want false")
	}
}

func TestEvaluateDistribution(t *testing.T) {
	const users = 10000
	const percent = 50

	var count int
	for i := 0; i < users; i++ {
		if Evaluate("feature", fmt.Sprintf("user-%d", i), percent, true) {
			count++
		}
	}

	ratio := float64(count) / float64(users)
	if ratio < 0.40 || ratio > 0.60 {
		t.Errorf("distribution too skewed: got %.2f%% over %d users, want ~%.2f%%", ratio*100, users, float64(percent))
	}
}
