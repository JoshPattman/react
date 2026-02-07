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
