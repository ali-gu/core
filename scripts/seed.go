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
	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/internal/storage/states"
	"github.com/ali-gulzar/speechory-core/pkg/ptr"
	"github.com/jackc/pgx/v5"
	"github.com/segmentio/ksuid"
)

var (
	seedPracticeID    = mustParseKSUID("3GateuqaP9ReKyB0Rwit6XJb9FP")
	seedUserID        = mustParseKSUID("3Gatex8E9CzpB43nDrM9dHzmrLD")
	seedLocationID    = mustParseKSUID("3H3QbRMyzFRnR3LuOLKVJG4Bt1H")
	seedEHRID         = mustParseKSUID("3H3QbRwpWXaTyQO7YgUQqD0X09M")
	seedPhoneNumberID = mustParseKSUID("3H3QbRrT1CVLUQfjx0tqlzSQxhG")
	seedAgentID       = mustParseKSUID("3H3QbWhEQUBZd5p2ZWuinJoWtM3")
)

const (
	seedUserEmail = "local@speechory.com"
	seedUserRef   = "4bce7cdb-8173-4322-8422-4b81b752e7ad"

	seedPracticeEmail   = "local@speechory"
	seedPracticeWebsite = "https://speechory.com"
	seedPracticeZipCode = "30101"

	seedLocationAddress = "5345 Magnolia"

	seedEHRSubdomain   = "speechory-demo-practice"
	seedEHRLocationRef = "353210"

	seedPhoneNumber    = "+13366541083"
	seedPhoneNumberRef = "3012822334093395781"

	seedAgentRef  = "assistant-cf1ab171-db2b-48b8-a7b6-57930f547615"
	seedAgentName = "Local Agent"
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

	if err := seed(ctx, db, storage.NewStorage()); err != nil {
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

func seed(ctx context.Context, db storage.DB, s storage.Storage) error {
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

	if err := s.Location.Create(ctx, db, storage.Location{
		EntityBase: storage.EntityBase[states.LocationState]{EntityState: states.LocationStateActive},
		ID:         seedLocationID,
		Address:    seedLocationAddress,
		PracticeID: seedPracticeID,
		CreatedAt:  time.Now(),
	}); err != nil {
		return fmt.Errorf("seed: creating location: %w", err)
	}

	onboardingID := ksuid.New().String()
	if err := s.EHR.Create(ctx, db, storage.EHRS{
		ID:            seedEHRID,
		Type:          constants.EHRNexHealth,
		Subdomain:     seedEHRSubdomain,
		LocationRef:   ptr.To(seedEHRLocationRef),
		LocationID:    seedLocationID,
		OnboardingURL: "https://app.nexhealth.com/onboardings/" + onboardingID,
		OnboardingID:  onboardingID,
		CreatedAt:     time.Now(),
	}); err != nil {
		return fmt.Errorf("seed: creating ehr: %w", err)
	}

	if err := s.PhoneNumber.Create(ctx, db, storage.PhoneNumber{
		EntityBase:       storage.EntityBase[states.PhoneNumberState]{EntityState: states.PhoneNumberStateActive},
		ID:               seedPhoneNumberID,
		PhoneNumber:      seedPhoneNumber,
		PhoneNumberIDRef: ptr.To(seedPhoneNumberRef),
		PracticeID:       seedPracticeID,
		CreatedAt:        time.Now(),
	}); err != nil {
		return fmt.Errorf("seed: creating phone number: %w", err)
	}

	if err := s.Agent.Create(ctx, db, storage.Agent{
		EntityBase:    storage.EntityBase[states.AgentState]{EntityState: states.AgentStateActive},
		ID:            seedAgentID,
		PracticeID:    seedPracticeID,
		AgentRef:      ptr.To(seedAgentRef),
		Name:          seedAgentName,
		LocationID:    ptr.To(seedLocationID),
		PhoneNumberID: ptr.To(seedPhoneNumberID),
		CreatedAt:     time.Now(),
	}); err != nil {
		return fmt.Errorf("seed: creating agent: %w", err)
	}

	log.Printf("seed: created location, ehr, phone number, and agent (practice %s)", seedPracticeID)
	return nil
}
