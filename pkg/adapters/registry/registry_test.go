package registry

import "testing"

func TestLookupMissingInput(t *testing.T) {
	if _, err := Input("http"); err == nil {
		t.Fatal("expected missing input error when build tag is absent")
	}
}

func TestLookupMissingSink(t *testing.T) {
	if _, err := Sink("mqtt"); err == nil {
		t.Fatal("expected missing sink error when build tag is absent")
	}
}
