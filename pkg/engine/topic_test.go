package engine

import "testing"

func TestBuildTopic(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		groupID   string
		msgType   string
		nodeID    string
		deviceID  string
		want      string
	}{
		{name: "node birth", namespace: "", groupID: "group", msgType: "NBIRTH", nodeID: "node", want: "spBv1.0/group/NBIRTH/node"},
		{name: "node data", namespace: "", groupID: "group", msgType: "NDATA", nodeID: "node", want: "spBv1.0/group/NDATA/node"},
		{name: "device birth", namespace: "", groupID: "group", msgType: "DBIRTH", nodeID: "node", deviceID: "device", want: "spBv1.0/group/DBIRTH/node/device"},
		{name: "device data", namespace: "", groupID: "group", msgType: "DDATA", nodeID: "node", deviceID: "device", want: "spBv1.0/group/DDATA/node/device"},
		{name: "device death", namespace: "", groupID: "group", msgType: "DDEATH", nodeID: "node", deviceID: "device", want: "spBv1.0/group/DDEATH/node/device"},
		{name: "node death", namespace: "", groupID: "group", msgType: "NDEATH", nodeID: "node", want: "spBv1.0/group/NDEATH/node"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildTopic(tt.namespace, tt.groupID, tt.msgType, tt.nodeID, tt.deviceID)
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}
