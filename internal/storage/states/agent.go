package states

type AgentState string

const (
	AgentStateCreated  AgentState = "CREATED"
	AgentStateActive   AgentState = "ACTIVE"
	AgentStateDisabled AgentState = "DISABLED"
)
