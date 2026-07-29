package constants

type ToolType string

const (
	ToolTypeWebhook   ToolType = "webhook"
	ToolTypeFunction  ToolType = "function"
	ToolTypeRetrieval ToolType = "retrieval"
	ToolTypeHandoff   ToolType = "handoff"
	ToolTypeHangup    ToolType = "hangup"
)

func (t ToolType) IsValid() bool {
	switch t {
	case ToolTypeWebhook, ToolTypeFunction, ToolTypeRetrieval, ToolTypeHandoff, ToolTypeHangup:
		return true
	default:
		return false
	}
}

type ToolKind string

const (
	ToolKindHangUp            ToolKind = "Default hangup" // default tool available
	ToolKindBookAppointment   ToolKind = "book_appointment"
	ToolKindFindAppointment   ToolKind = "find_appointment"
	ToolKindCancelAppointment ToolKind = "cancel_appointment"
	ToolKindLogConversation   ToolKind = "log_conversation"
)

func (k ToolKind) IsValid() bool {
	switch k {
	case ToolKindBookAppointment, ToolKindFindAppointment, ToolKindCancelAppointment, ToolKindHangUp, ToolKindLogConversation:
		return true
	default:
		return false
	}
}

func (k ToolKind) String() string {
	return string(k)
}
