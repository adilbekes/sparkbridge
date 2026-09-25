package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucketCapacityAndRefill(t *testing.T) {
	bucket := New(2, 10*time.Millisecond)
	if !bucket.Allow() || !bucket.Allow() {
		t.Fatal("expected initial capacity")
	}
	if bucket.Allow() {
		t.Fatal("expected exhausted bucket")
	}
	time.Sleep(15 * time.Millisecond)
	if !bucket.Allow() {
		t.Fatal("expected refilled token")
	}
}

func TestTokenBucketNormalizesCapacity(t *testing.T) {
	bucket := New(0, 0)
	if !bucket.Allow() {
		t.Fatal("expected minimum capacity of one")
	}
	if bucket.Allow() {
		t.Fatal("expected bucket to be exhausted")
	}
}
