package pipeline

import "time"

// EncodedMessage is the output of the pipeline worker pool.
type EncodedMessage struct {
	Topic     string
	Payload   []byte
	Timestamp time.Time
}
