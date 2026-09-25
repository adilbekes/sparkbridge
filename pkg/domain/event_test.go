package domain

import "testing"

func TestEventCloneIsIndependent(t *testing.T) {
	alias := uint64(7)
	original := Event{
		Metrics:    []Metric{{Name: "bytes", Value: []byte{1, 2}, Alias: &alias, Properties: map[string]any{"nested": map[string]any{"value": "original"}}}},
		Properties: map[string]any{"payload": []byte{3, 4}},
	}
	clone := original.Clone()
	clone.Metrics[0].Value.([]byte)[0] = 9
	*clone.Metrics[0].Alias = 8
	clone.Metrics[0].Properties["nested"].(map[string]any)["value"] = "changed"
	clone.Properties["payload"].([]byte)[0] = 9

	if original.Metrics[0].Value.([]byte)[0] != 1 || *original.Metrics[0].Alias != 7 {
		t.Fatal("metric clone mutated original")
	}
	if original.Metrics[0].Properties["nested"].(map[string]any)["value"] != "original" {
		t.Fatal("nested metric properties mutated original")
	}
	if original.Properties["payload"].([]byte)[0] != 3 {
		t.Fatal("event properties mutated original")
	}
}
