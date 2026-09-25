//go:build input_mqtt || all_inputs

package mqtt

import "testing"

func TestDecode(t *testing.T) {
	if _, err := New().Decode("sensors/x/telemetry", []byte(`{"group_id":"g","node_id":"n","msg_type":"NDATA","metrics":[{"name":"temp","value":1}]}`)); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
