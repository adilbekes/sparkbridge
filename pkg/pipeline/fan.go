package pipeline

import "sparkbridge/pkg/domain"

// FanIn multiplexes inputs into the pipeline.
func FanIn(inputs ...<-chan domain.Event) <-chan domain.Event {
	out := make(chan domain.Event)
	go func() {
		defer close(out)
		for _, ch := range inputs {
			for evt := range ch {
				out <- evt
			}
		}
	}()
	return out
}

// FanOut duplicates messages to multiple sinks.
func FanOut(input <-chan EncodedMessage, sinks ...chan<- EncodedMessage) {
	go func() {
		for msg := range input {
			for _, sink := range sinks {
				sink <- msg
			}
		}
		for _, sink := range sinks {
			close(sink)
		}
	}()
}
