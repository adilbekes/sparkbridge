package engine

import "sparkbridge/pkg/domain"

// BuildTopic builds a Sparkplug topic using the default namespace when empty.
func BuildTopic(namespace, groupID, msgType, nodeID, deviceID string) string {
	if namespace == "" {
		namespace = "spBv1.0"
	}
	if deviceID != "" {
		return namespace + "/" + groupID + "/" + msgType + "/" + nodeID + "/" + deviceID
	}
	return namespace + "/" + groupID + "/" + msgType + "/" + nodeID
}

// TopicForEvent derives the topic from a domain event.
func TopicForEvent(event domain.Event) string {
	return BuildTopic("", event.GroupID, string(event.MsgType), event.NodeID, event.DeviceID)
}
