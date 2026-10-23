package jwtoggl

import (
	"testing"
	"time"
)

func TestSyncDelayForErrorsUsesNormalIntervalAfterSuccess(t *testing.T) {
	if got := syncDelayForErrors(0); got != normalSyncInterval {
		t.Fatalf("syncDelayForErrors(0) = %s, want %s", got, normalSyncInterval)
	}
	if normalSyncInterval < 2*time.Second {
		t.Fatalf("normalSyncInterval = %s, want at least 2s for two REST reads", normalSyncInterval)
	}
}

func TestSyncDelayForErrorsKeepsIncrementalBackoff(t *testing.T) {
	if got, want := syncDelayForErrors(1), 15*time.Minute; got != want {
		t.Fatalf("syncDelayForErrors(1) = %s, want %s", got, want)
	}
}

func TestSyncDelayForFailuresUsesQuotaReset(t *testing.T) {
	if got, want := syncDelayForFailures(2, true, 37*time.Second), 37*time.Second; got != want {
		t.Fatalf("syncDelayForFailures() = %s, want %s", got, want)
	}
}

func TestSyncDelayForFailuresUsesSafeFallbackForQuotaWithoutHeader(t *testing.T) {
	if got, want := syncDelayForFailures(0, true, 0), time.Hour; got != want {
		t.Fatalf("syncDelayForFailures() = %s, want %s", got, want)
	}
}
