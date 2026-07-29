package storage_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

func newPooledMux(t *testing.T) *storage.DBMux {
	t.Helper()

	cfg, err := internal.NewConfig(constants.LandscapeTest)
	require.NoError(t, err)

	mux, err := storage.NewDBMux(context.Background(), *cfg)
	require.NoError(t, err)
	t.Cleanup(mux.Close)

	return mux
}

func Test_Storage_ConcurrentReads_DoNotDeadlockOrRace(t *testing.T) {
	mux := newPooledMux(t)
	db, err := mux.BeginRead()
	require.NoError(t, err)

	store := storage.NewStorage()
	roleTypes := []constants.RoleType{
		constants.RoleTypeSuperAdmin,
		constants.RoleTypeAdmin,
		constants.RoleTypeReader,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const goroutines = 64
	const iterations = 25

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*iterations)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				want := roleTypes[(g+i)%len(roleTypes)]
				roles, err := store.Role.Get(ctx, db, storage.RoleFilters{Type: &want})
				if err != nil {
					errCh <- fmt.Errorf("get role %s: %w", want, err)
					return
				}
				if len(roles) != 1 || roles[0].Type != want {
					errCh <- fmt.Errorf("concurrent read returned wrong role: want %s, got %+v", want, roles)
					return
				}
			}
		}(g)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}
}

func Test_Storage_ConcurrentWrites_ExceedPoolSizeWithoutDeadlock(t *testing.T) {
	mux := newPooledMux(t)
	db, err := mux.BeginWrite()
	require.NoError(t, err)

	store := storage.NewStorage()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const goroutines = 40

	ids := make([]string, goroutines)
	for i := range ids {
		ids[i] = ksuid.New().String()
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cleanupCancel()
		_, _ = db.Exec(cleanupCtx, "DELETE FROM practices WHERE id = ANY($1)", ids)
	})

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			id, parseErr := ksuid.Parse(ids[i])
			if parseErr != nil {
				errCh <- fmt.Errorf("parse id: %w", parseErr)
				return
			}

			practice := storage.Practice{
				EntityBase: storage.EntityBase[states.PracticeState]{EntityState: states.PracticeStateCreated},
				ID:         id,
				Name:       fmt.Sprintf("conc_practice_%d", i),
				CreatedAt:  time.Now(),
			}
			if createErr := store.Practice.Create(ctx, db, practice); createErr != nil {
				errCh <- fmt.Errorf("create practice %d: %w", i, createErr)
				return
			}

			fetched, getErr := store.Practice.GetByID(ctx, db, id)
			if getErr != nil {
				errCh <- fmt.Errorf("get practice %d: %w", i, getErr)
				return
			}
			if fetched.ID != id {
				errCh <- fmt.Errorf("fetched wrong practice: want %s got %s", id, fetched.ID)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		require.NoError(t, err)
	}

	for _, rawID := range ids {
		id, parseErr := ksuid.Parse(rawID)
		require.NoError(t, parseErr)
		fetched, getErr := store.Practice.GetByID(ctx, db, id)
		require.NoError(t, getErr)
		require.Equal(t, id, fetched.ID)
	}
}
