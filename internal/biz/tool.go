package biz

import (
	"context"
	"errors"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/services/tool"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type Tool struct {
	*Biz

	storage    storage.Storage
	telnyxTool tool.ITool
	domain     string
}

type ITool interface {
	Sync(ctx context.Context, db storage.DB) ([]storage.Tool, error)
}

var _ ITool = (*Tool)(nil)

func (b *Tool) Sync(ctx context.Context, db storage.DB) ([]storage.Tool, error) {
	remoteTools, err := b.telnyxTool.List(ctx)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	for _, remote := range remoteTools {
		kind := constants.ToolKind(remote.DisplayName)
		if !kind.IsValid() {
			continue
		}

		existing, err := b.storage.Tool.GetByKind(ctx, db, kind)
		switch {
		case err == nil:
			existing.ToolRef = remote.ID
			existing.Config = remote.Config
			if err := b.storage.Tool.Update(ctx, db, *existing); err != nil {
				return nil, rerror.Wrap(err)
			}
		case errors.Is(err, pgx.ErrNoRows):
			if err := b.storage.Tool.Create(ctx, db, storage.Tool{
				EntityBase: storage.EntityBase[states.ToolState]{EntityState: states.ToolStateActive},
				ID:         ksuid.New(),
				Type:       constants.ToolType(remote.Type),
				Kind:       kind,
				ToolRef:    remote.ID,
				Config:     remote.Config,
				CreatedAt:  time.Now(),
			}); err != nil {
				return nil, rerror.Wrap(err)
			}
		default:
			return nil, rerror.Wrap(err)
		}
	}

	return b.storage.Tool.Get(ctx, db, storage.ToolFilters{})
}
