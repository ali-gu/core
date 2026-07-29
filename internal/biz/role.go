package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

type Role struct {
	*Biz

	storage storage.Storage
}

type IRole interface {
	Create(ctx context.Context, db storage.DB, input contracts.CreateRoleRequest) (*storage.Role, error)
}

var _ IRole = (*Role)(nil)

func (r *Role) Create(ctx context.Context, db storage.DB, input contracts.CreateRoleRequest) (*storage.Role, error) {
	if !input.Type.IsValid() {
		return nil, rerror.NewMessage(fmt.Sprintf("invalid role type %q", input.Type), rerror.Validation)
	}

	if _, err := r.storage.Role.GetByType(ctx, db, input.Type); err == nil {
		return nil, rerror.NewMessage(fmt.Sprintf("role of type %q already exists", input.Type), rerror.Validation)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, rerror.Wrap(err)
	}

	id := ksuid.New()

	role := storage.Role{
		EntityBase: storage.EntityBase[states.RoleState]{EntityState: states.RoleStateActive},
		ID:         id,
		Type:       input.Type,
		CreatedAt:  time.Now(),
	}
	if err := r.storage.Role.Create(ctx, db, role); err != nil {
		return nil, rerror.Wrap(err)
	}

	return r.storage.Role.GetByID(ctx, db, id)
}
