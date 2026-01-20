package react

import (
	"iter"

	"github.com/JoshPattman/jpf"
)

// Agent defines a Re-Act Agent.
type Agent interface {
	// Send a message to the agent, getting its response.
	// May pass parameters to stream back messages or the response.
	Send(msg string, opts ...SendMessageOpt) (string, error)
	// Fetch the entire history state of the agent.
	Messages() iter.Seq[Message]
}

// MessageStreamer defines a callback interface that can be used to listen to new messages that the agent creates.
type MessageStreamer interface {
	// Try to send a message, ignoring errors.
	TrySendMessage(msg Message)
}

// TextStreamer defines a callback interface that can be used to listen to text being streamed back from the final agent response.
type TextStreamer interface {
	// Try to send a response text chunk back, ignoring errors.
	TrySendTextChunk(chunk string)
}

type SendMessageOpt func(*streamers)

// In addition to other message streamers, use the provided streamer.
func WithMessageStreamer(streamer MessageStreamer) SendMessageOpt {
	return func(s *streamers) {
		s.msgStreamers = append(s.msgStreamers, streamer)
	}
}

// In addition to other response streamers, use the provided streamer.
func WithResponseStreamer(streamer TextStreamer) SendMessageOpt {
	return func(s *streamers) {
		s.respStreamers = append(s.respStreamers, streamer)
	}
}

type ModelBuilder interface {
	BuildAgentModel(responseType any, onInitFinalStream func(), onDataFinalStream func(string)) jpf.Model
}

type streamers struct {
	msgStreamers  []MessageStreamer
	respStreamers []TextStreamer
}

func getStreamers(opts []SendMessageOpt) streamers {
	s := streamers{}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

func (s streamers) TrySendMessage(msg Message) {
	for _, msgStreamer := range s.msgStreamers {
		msgStreamer.TrySendMessage(msg)
	}
}

func (s streamers) TrySendTextChunk(chunk string) {
	for _, msgStreamer := range s.respStreamers {
		msgStreamer.TrySendTextChunk(chunk)
	}
}
