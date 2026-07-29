package telnyx_test

import (
	"testing"

	"github.com/ali-gulzar/speechory-core/internal/services/agent/telnyx"
	"github.com/stretchr/testify/require"
)

func Test_ConversationOwnedByAgent(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]string
		agentRef string
		owned    bool
	}{
		{
			name:     "matching_assistant_id_is_owned",
			metadata: map[string]string{"assistant_id": "assistant-123"},
			agentRef: "assistant-123",
			owned:    true,
		},
		{
			name:     "mismatched_assistant_id_is_not_owned",
			metadata: map[string]string{"assistant_id": "assistant-999"},
			agentRef: "assistant-123",
			owned:    false,
		},
		{
			name:     "missing_assistant_id_is_not_owned",
			metadata: map[string]string{"call_control_id": "cc-1"},
			agentRef: "assistant-123",
			owned:    false,
		},
		{
			name:     "nil_metadata_is_not_owned",
			metadata: nil,
			agentRef: "assistant-123",
			owned:    false,
		},
		{
			name:     "empty_agent_ref_is_not_owned",
			metadata: map[string]string{"assistant_id": "assistant-123"},
			agentRef: "",
			owned:    false,
		},
		{
			name:     "empty_agent_ref_and_empty_assistant_id_is_not_owned",
			metadata: map[string]string{"assistant_id": ""},
			agentRef: "",
			owned:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.owned, telnyx.ConversationOwnedByAgent(tt.metadata, tt.agentRef))
		})
	}
}
