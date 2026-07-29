package telnyx

import (
	"context"
	"fmt"

	"github.com/ali-gulzar/speechory-core/internal/services/tool"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"

	telnyxsdk "github.com/team-telnyx/telnyx-go/v4"
)

type Telnyx struct {
	client telnyxsdk.Client
}

var _ tool.ITool = (*Telnyx)(nil)

func NewTelnyx(client telnyxsdk.Client) *Telnyx {
	return &Telnyx{client: client}
}

func (t *Telnyx) List(ctx context.Context) ([]tool.ListToolsResult, error) {
	var results []tool.ListToolsResult

	iter := t.client.AI.Tools.ListAutoPaging(ctx, telnyxsdk.AIToolListParams{})
	for iter.Next() {
		item := iter.Current()
		results = append(results, tool.ListToolsResult{
			ID:          item.ID,
			Type:        item.Type,
			DisplayName: item.DisplayName,
			Config:      item.ToolDefinition,
		})
	}
	if err := iter.Err(); err != nil {
		return nil, rerror.New(fmt.Errorf("telnyx: list tools: %w", err))
	}

	return results, nil
}
