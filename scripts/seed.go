package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	cfgfiles "github.com/ali-gulzar/speechory-core/config"
	"github.com/ali-gulzar/speechory-core/internal"
	"github.com/ali-gulzar/speechory-core/internal/biz"
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

var (
	seedPracticeID = mustParseKSUID("3GateuqaP9ReKyB0Rwit6XJb9FP")
	seedUserID     = mustParseKSUID("3Gatex8E9CzpB43nDrM9dHzmrLD")
)

const (
	seedUserEmail = "local@speechory.com"
	seedUserRef   = "4bce7cdb-8173-4322-8422-4b81b752e7ad"

	seedPracticeEmail   = "local@speechory"
	seedPracticeWebsite = "https://speechory.com"
	seedPracticeZipCode = "30101"
)

func mustParseKSUID(s string) ksuid.KSUID {
	id, err := ksuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

func main() {
	ctx := context.Background()

	dbConfig, err := loadDBConfig()
	if err != nil {
		log.Fatal(err)
	}

	dbMux, err := storage.NewDBMux(ctx, internal.Config{DB: dbConfig})
	if err != nil {
		log.Fatal(err)
	}
	defer dbMux.Close()

	db, err := dbMux.BeginWrite()
	if err != nil {
		log.Fatal(err)
	}

	storageDeps := storage.NewStorage()
	globalBiz := biz.NewBiz(biz.Dependencies{
		Storage:   storageDeps,
		Landscape: constants.LandscapeLocal,
	})

	if err := seed(ctx, db, storageDeps, globalBiz); err != nil {
		log.Fatal(err)
	}

	log.Println("seed: done")
}

func loadDBConfig() (internal.DBConfig, error) {
	raw, err := cfgfiles.Files.ReadFile("local.json")
	if err != nil {
		return internal.DBConfig{}, err
	}

	var wrapper struct {
		DB internal.DBConfig `json:"db"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return internal.DBConfig{}, err
	}

	return wrapper.DB, nil
}

func seed(ctx context.Context, db storage.DB, s storage.Storage, bz *biz.Biz) error {
	if _, err := s.User.GetByEmail(ctx, db, seedUserEmail); err == nil {
		log.Printf("seed: user %s already exists, skipping", seedUserEmail)
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("seed: checking for existing seed user: %w", err)
	}

	if err := s.Practice.Create(ctx, db, storage.Practice{
		EntityBase: storage.EntityBase[states.PracticeState]{EntityState: states.PracticeStateActive},
		ID:         seedPracticeID,
		Name:       "Speechory",
		Email:      ptr.To(seedPracticeEmail),
		Website:    ptr.To(seedPracticeWebsite),
		ZipCode:    ptr.To(seedPracticeZipCode),
		CreatedAt:  time.Now(),
	}); err != nil {
		return fmt.Errorf("seed: creating practice: %w", err)
	}

	roleType := constants.RoleTypeSuperAdmin
	role, err := s.Role.GetByType(ctx, db, roleType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("seed: no active %s role found", roleType)
		}
		return fmt.Errorf("seed: looking up %s role: %w", roleType, err)
	}
	if role.EntityState != states.RoleStateActive {
		return fmt.Errorf("seed: no active %s role found", roleType)
	}

	if err := s.User.Create(ctx, db, storage.User{
		EntityBase: storage.EntityBase[states.UserState]{EntityState: states.UserStateActive},
		ID:         seedUserID,
		UserRef:    seedUserRef,
		RoleID:     role.ID,
		PracticeID: seedPracticeID,
		Email:      seedUserEmail,
		CreatedAt:  time.Now(),
	}); err != nil {
		return fmt.Errorf("seed: creating user: %w", err)
	}

	log.Printf("seed: created user %s (practice %s)", seedUserEmail, seedPracticeID)

	testAgent, err := bz.TestAgent.Create(ctx, db, seedPracticeID)
	if err != nil {
		return fmt.Errorf("seed: creating test agent: %w", err)
	}

	log.Printf("seed: created location, ehr, phone number, and agent %s (practice %s)", testAgent.ID, seedPracticeID)
	return nil
}
