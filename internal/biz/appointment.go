package biz

import (
	"context"
	"fmt"

	"github.com/ali-gulzar/speechory-core/internal/constants"
	"github.com/ali-gulzar/speechory-core/internal/contracts"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/storage"
	"github.com/ali-gulzar/speechory-core/pkg/rerror"
)

type Appointment struct {
	*Biz

	storage storage.Storage
	ehr     ehr.IEHR
}

type IAppointment interface {
	Find(ctx context.Context, db storage.DB, input contracts.FindAppointmentRequest) ([]ehr.Appointment, error)
	Book(ctx context.Context, db storage.DB, input contracts.BookAppointmentRequest) (*ehr.BookedAppointment, error)
	Cancel(ctx context.Context, db storage.DB, input contracts.CancelAppointmentRequest) (*ehr.CancelledAppointment, error)
}

var _ IAppointment = (*Appointment)(nil)

func (a *Appointment) Find(ctx context.Context, db storage.DB, input contracts.FindAppointmentRequest) ([]ehr.Appointment, error) {
	subdomain, locationRef, err := a.resolveLocationEHRByAgentRef(ctx, db, input.AssistantID)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	return a.ehr.FindAppointment(ctx, ehr.AvailableAppointmentsParams{
		Subdomain:  subdomain,
		LocationID: locationRef,
		StartDate:  input.StartDate,
		Days:       constants.DefaultAppointmentSearchDays,
	})
}

func (a *Appointment) Book(ctx context.Context, db storage.DB, input contracts.BookAppointmentRequest) (*ehr.BookedAppointment, error) {
	subdomain, locationRef, err := a.resolveLocationEHRByAgentRef(ctx, db, input.AssistantID)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	return a.ehr.BookAppointment(ctx, ehr.BookAppointmentParams{
		Subdomain:  subdomain,
		LocationID: locationRef,
		ProviderID: input.DoctorID,
		Time:       input.Time,
		Note:       input.Reason,
		Patient: ehr.PatientDetails{
			FirstName:   input.Patient.FirstName,
			LastName:    input.Patient.LastName,
			Phone:       input.Patient.Phone,
			DateOfBirth: input.Patient.DateOfBirth,
		},
	})
}

func (a *Appointment) Cancel(ctx context.Context, db storage.DB, input contracts.CancelAppointmentRequest) (*ehr.CancelledAppointment, error) {
	subdomain, locationRef, err := a.resolveLocationEHRByAgentRef(ctx, db, input.AssistantID)
	if err != nil {
		return nil, rerror.Wrap(err)
	}

	return a.ehr.CancelAppointment(ctx, ehr.CancelAppointmentParams{
		Subdomain:  subdomain,
		LocationID: locationRef,
		Patient: ehr.PatientLookup{
			Name:        input.Patient.Name,
			Phone:       input.Patient.Phone,
			DateOfBirth: input.Patient.DateOfBirth,
		},
	})
}

func (a *Appointment) resolveLocationEHRByAgentRef(ctx context.Context, db storage.DB, assistantID string) (subdomain string, locationRef string, err error) {
	agentRecord, err := a.storage.Agent.GetByAgentRef(ctx, db, assistantID)
	if err != nil {
		return "", "", rerror.Wrap(err)
	}

	return a.resolveLocationEHRForAgent(ctx, db, agentRecord)
}

func (a *Appointment) resolveLocationEHRForAgent(ctx context.Context, db storage.DB, agentRecord *storage.Agent) (subdomain string, locationRef string, err error) {
	if agentRecord.LocationID == nil {
		return "", "", rerror.NewMessage(fmt.Sprintf("agent %s has no location assigned", agentRecord.ID), rerror.Validation)
	}
	location, err := a.storage.Location.GetByID(ctx, db, *agentRecord.LocationID)
	if err != nil {
		return "", "", rerror.Wrap(err)
	}

	e, err := a.storage.EHR.GetByLocationID(ctx, db, location.ID)
	if err != nil {
		return "", "", rerror.Wrap(err)
	}

	if e.Subdomain == "" || e.LocationRef == nil || *e.LocationRef == "" {
		return "", "", rerror.New(fmt.Errorf("biz: ehr %s is missing nexhealth configuration", e.ID))
	}

	return e.Subdomain, *e.LocationRef, nil
}
