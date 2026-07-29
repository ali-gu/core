package biz

import (
	"context"
	"time"

	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
	"github.com/segmentio/ksuid"
)

type Practice struct {
	*Biz

	storage storage.Storage
}

type IPractice interface {
	Create(ctx context.Context, db storage.DB, input contracts.CreatePracticeRequest) (*storage.Practice, error)
	GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Practice, error)
	Update(ctx context.Context, db storage.DB, id ksuid.KSUID, input contracts.UpdatePracticeRequest) (*storage.Practice, error)
}

var _ IPractice = (*Practice)(nil)

func (p *Practice) Create(ctx context.Context, db storage.DB, input contracts.CreatePracticeRequest) (*storage.Practice, error) {
	id := ksuid.New()

	practice := storage.Practice{
		EntityBase: storage.EntityBase[states.PracticeState]{EntityState: states.PracticeStateCreated},
		ID:         id,
		Name:       input.Name,
		Email:      input.Email,
		ZipCode:    input.ZipCode,
		Website:    input.Website,
		CreatedAt:  time.Now(),
	}
	if err := p.storage.Practice.Create(ctx, db, practice); err != nil {
		return nil, rerror.Wrap(err)
	}

	result, err := p.storage.Practice.GetByID(ctx, db, id)
	if err != nil {
		return nil, rerror.Wrap(err)
	}
	return result, nil
}

func (p *Practice) GetByID(ctx context.Context, db storage.DB, id ksuid.KSUID) (*storage.Practice, error) {
	return p.storage.Practice.GetByID(ctx, db, id)
}

func (p *Practice) Update(ctx context.Context, db storage.DB, id ksuid.KSUID, input contracts.UpdatePracticeRequest) (*storage.Practice, error) {
	practice, err := p.storage.Practice.GetByID(ctx, db, id)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	if input.Email != nil {
		practice.Email = input.Email
	}
	if input.ZipCode != nil {
		practice.ZipCode = input.ZipCode
	}
	if input.Website != nil {
		practice.Website = input.Website
	}

	if err := p.storage.Practice.Update(ctx, db, *practice); err != nil {
		return nil, rerror.Wrap(err)
	}

	return p.storage.Practice.GetByID(ctx, db, id)
}
