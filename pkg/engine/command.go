package engine

import "sparkbridge/pkg/domain"

// IsRebirthCommand reports whether the event requests a node rebirth.
func IsRebirthCommand(event domain.Event) bool {
	if event.MsgType != domain.MessageTypeNCMD {
		return false
	}
	for _, metric := range event.Metrics {
		if metric.Name == "Node Control/Rebirth" {
			if value, ok := metric.Value.(bool); ok {
				return value
			}
			if value, ok := metric.Value.(string); ok {
				return value == "true"
			}
		}
	}
	return false
}
