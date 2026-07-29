package tool

import (
	"context"
)

type ITool interface {
	List(ctx context.Context) ([]ListToolsResult, error)
}
