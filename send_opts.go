package react

type SendMessageOpt func(*sendMessageKwargs)

// In addition to other message streamers, use the provided streamer.
func WithMessageStreamer(streamer MessageStreamer) SendMessageOpt {
	return func(s *sendMessageKwargs) {
		s.msgStreamers = append(s.msgStreamers, streamer)
	}
}

// In addition to other response streamers, use the provided streamer.
func WithResponseStreamer(streamer TextStreamer) SendMessageOpt {
	return func(s *sendMessageKwargs) {
		s.respStreamers = append(s.respStreamers, streamer)
	}
}

func WithNotifications(notifications ...Notification) SendMessageOpt {
	return func(s *sendMessageKwargs) {
		s.notifications = append(s.notifications, notifications...)
	}
}

type sendMessageKwargs struct {
	msgStreamers  []MessageStreamer
	respStreamers []TextStreamer
	notifications []Notification
}

func getKwargs(opts []SendMessageOpt) sendMessageKwargs {
	s := sendMessageKwargs{}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

func (kw sendMessageKwargs) Streamers() multiStreamers {
	return multiStreamers{kw.msgStreamers, kw.respStreamers}
}
