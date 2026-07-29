package storage

import "github.com/Masterminds/squirrel"

var StatementBuilder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type Storage struct {
	Agent         IAgentStorage
	Practice      IPracticeStorage
	Location      ILocationStorage
	PhoneNumber   IPhoneNumberStorage
	EHR           IEHRStorage
	Role          IRoleStorage
	User          IUserStorage
	Tool          IToolStorage
	Conversation  IConversationStorage
	NexHealthAuth INexHealthAuthStorage
	HTTPRequest   IHTTPRequestStorage
}

func NewStorage() Storage {
	return Storage{
		Agent:         &AgentStorage{},
		Practice:      &PracticeStorage{},
		Location:      &LocationStorage{},
		PhoneNumber:   &PhoneNumberStorage{},
		EHR:           &EHRStorage{},
		Role:          &RoleStorage{},
		User:          &UserStorage{},
		Tool:          &ToolStorage{},
		Conversation:  &ConversationStorage{},
		NexHealthAuth: &NexHealthAuthStorage{},
		HTTPRequest:   &HTTPRequestStorage{},
	}
}

type EntityBase[T any] struct {
	EntityState T `json:"entity_state" db:"entity_state"`
}
