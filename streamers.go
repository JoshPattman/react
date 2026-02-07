package react

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

type multiStreamers struct {
	msgStreamers  []MessageStreamer
	respStreamers []TextStreamer
}

func (s multiStreamers) TrySendMessage(msg Message) {
	for _, msgStreamer := range s.msgStreamers {
		msgStreamer.TrySendMessage(msg)
	}
}

func (s multiStreamers) TrySendTextChunk(chunk string) {
	for _, msgStreamer := range s.respStreamers {
		msgStreamer.TrySendTextChunk(chunk)
	}
}
