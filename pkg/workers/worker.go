package workers

type Worker func(string, chan Message)

const (
	MessageTypeProgress = iota
	MessageTypeInfo
	MessageTypeError
	MessageTypeTitle
)

type Message struct {
	Type    int8
	Content string
}

func info_msg(text string) Message {
	return Message{Type: MessageTypeInfo, Content: text}
}

func error_msg(err string) Message {
	return Message{Type: MessageTypeError, Content: err}
}
