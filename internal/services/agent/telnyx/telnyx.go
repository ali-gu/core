package telnyx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"

	telnyxsdk "github.com/team-telnyx/telnyx-go/v4"
	"github.com/team-telnyx/telnyx-go/v4/packages/param"
)

type Telnyx struct {
	client    telnyxsdk.Client
	landscape constants.Landscape
}

var _ agent.IAgent = (*Telnyx)(nil)

func NewTelnyx(client telnyxsdk.Client, landscape constants.Landscape) *Telnyx {
	return &Telnyx{client: client, landscape: landscape}
}

func (t *Telnyx) Create(ctx context.Context, params agent.CreateAgentParams) (*agent.CreateAgentResult, error) {
	assistant, err := t.client.AI.Assistants.New(ctx, telnyxsdk.AIAssistantNewParams{
		Name:         params.Name,
		Instructions: params.Instructions,
		Model:        telnyxsdk.String(params.Model.String()),
		Greeting:     telnyxsdk.String(params.Greeting),
		ToolIDs:      params.ToolIDs,
		Tags:         []string{string(t.landscape)},
		PostConversationSettings: telnyxsdk.AIAssistantNewParamsPostConversationSettings{
			Enabled: param.NewOpt(true),
		},
	})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("telnyx: create agent: %w", err))
	}

	if params.PhoneNumberIDRef != "" {
		if _, err = t.client.PhoneNumbers.Update(ctx, params.PhoneNumberIDRef, telnyxsdk.PhoneNumberUpdateParams{
			ConnectionID: telnyxsdk.String(assistant.TelephonySettings.DefaultTexmlAppID),
		}); err != nil {
			return nil, rerror.New(fmt.Errorf("telnyx: assign phone number to agent: %w", err))
		}
	}

	return &agent.CreateAgentResult{
		ID:           assistant.ID,
		Name:         assistant.Name,
		Model:        constants.TelnyxModel(assistant.Model),
		Instructions: assistant.Instructions,
		CreatedAt:    assistant.CreatedAt,
	}, nil
}

func (t *Telnyx) Delete(ctx context.Context, params agent.DeleteAgentParams) error {
	if _, err := t.client.AI.Assistants.Delete(ctx, params.AgentRef); err != nil {
		return rerror.New(fmt.Errorf("telnyx: delete agent: %w", err))
	}
	return nil
}

func (t *Telnyx) GetAnalytics(ctx context.Context, params agent.GetAnalyticsParams) ([]agent.ConversationAnalytics, error) {
	conversations, err := t.client.AI.Conversations.List(ctx, telnyxsdk.AIConversationListParams{
		MetadataAssistantID: telnyxsdk.String("eq." + params.AgentRef),
	})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("telnyx: list conversations: %w", err))
	}

	results := make([]agent.ConversationAnalytics, 0, len(conversations.Data))
	for _, conversation := range conversations.Data {
		detail, err := t.buildConversationDetail(ctx, conversation)
		if err != nil {
			return nil, err
		}

		recordings, err := t.GetRecordings(ctx, agent.GetRecordingsParams{CallControlID: detail.Metadata["call_control_id"]})
		if err != nil {
			return nil, err
		}
		detail.Recordings = recordings

		results = append(results, detail)
	}

	return results, nil
}

func (t *Telnyx) GetConversation(ctx context.Context, params agent.GetConversationParams) (*agent.ConversationAnalytics, error) {
	list, err := t.client.AI.Conversations.List(ctx, telnyxsdk.AIConversationListParams{
		ID: telnyxsdk.String("eq." + params.ConversationID),
	})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("telnyx: get conversation: %w", err))
	}
	if len(list.Data) == 0 {
		return nil, rerror.New(agent.ErrConversationNotFound)
	}
	conversation := list.Data[0]

	detail, err := t.buildConversationDetail(ctx, conversation)
	if err != nil {
		return nil, err
	}

	if !ConversationOwnedByAgent(detail.Metadata, params.AgentRef) {
		return nil, rerror.New(agent.ErrConversationNotFound)
	}

	return &detail, nil
}

func ConversationOwnedByAgent(metadata map[string]string, agentRef string) bool {
	if agentRef == "" {
		return false
	}
	assistantID, ok := metadata["assistant_id"]
	return ok && assistantID == agentRef
}

func (t *Telnyx) buildConversationDetail(ctx context.Context, conversation telnyxsdk.Conversation) (agent.ConversationAnalytics, error) {
	insights, err := t.client.AI.Conversations.GetConversationsInsights(ctx, conversation.ID)
	if err != nil {
		return agent.ConversationAnalytics{}, rerror.New(fmt.Errorf("telnyx: get conversation insights: %w", err))
	}

	var status string
	insightData := make(map[string]string)
	for _, insight := range insights.Data {
		status = insight.Status
		for _, ci := range insight.ConversationInsights {
			insightData[ci.InsightID] = ci.Result
		}
	}

	messages, err := t.client.AI.Conversations.Messages.List(ctx, conversation.ID, telnyxsdk.AIConversationMessageListParams{})
	if err != nil {
		return agent.ConversationAnalytics{}, rerror.New(fmt.Errorf("telnyx: list conversation messages: %w", err))
	}

	conversationMessages := make([]agent.ConversationMessage, len(messages.Data))
	for i, message := range messages.Data {
		toolCalls := make([]agent.ConversationToolCall, len(message.ToolCalls))
		for j, toolCall := range message.ToolCalls {
			toolCalls[j] = agent.ConversationToolCall{
				ID:        toolCall.ID,
				Name:      toolCall.Function.Name,
				Arguments: toolCall.Function.Arguments,
			}
		}

		conversationMessages[i] = agent.ConversationMessage{
			Role:      string(message.Role),
			Text:      message.Text,
			SentAt:    message.SentAt,
			ToolCalls: toolCalls,
		}
	}

	return agent.ConversationAnalytics{
		ConversationID: conversation.ID,
		Status:         status,
		CreatedAt:      conversation.CreatedAt,
		LastMessageAt:  conversation.LastMessageAt,
		Metadata:       parseConversationMetadata(conversation.RawJSON()),
		Insights:       insightData,
		Messages:       conversationMessages,
	}, nil
}

func parseConversationMetadata(rawJSON string) map[string]string {
	var wrapper struct {
		Metadata map[string]json.RawMessage `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &wrapper); err != nil {
		return nil
	}

	metadata := make(map[string]string, len(wrapper.Metadata))
	for key, raw := range wrapper.Metadata {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			metadata[key] = s
			continue
		}
		metadata[key] = string(raw)
	}
	return metadata
}

func (t *Telnyx) GetRecordings(ctx context.Context, params agent.GetRecordingsParams) ([]agent.ConversationRecording, error) {
	if params.CallControlID == "" {
		return nil, nil
	}

	recordings, err := t.client.Recordings.List(ctx, telnyxsdk.RecordingListParams{
		Filter: telnyxsdk.RecordingListParamsFilter{
			CallControlID: telnyxsdk.String(params.CallControlID),
		},
	})
	if err != nil {
		return nil, rerror.New(fmt.Errorf("telnyx: list call recordings: %w", err))
	}

	results := make([]agent.ConversationRecording, len(recordings.Data))
	for i, recording := range recordings.Data {
		results[i] = agent.ConversationRecording{
			ID:             recording.ID,
			DurationMillis: recording.DurationMillis,
			MP3URL:         recording.DownloadURLs.MP3,
			WavURL:         recording.DownloadURLs.Wav,
			StartedAt:      recording.RecordingStartedAt,
			EndedAt:        recording.RecordingEndedAt,
		}
	}

	return results, nil
}
