package ehr

import "context"

type IEHR interface {
	Authenticate(ctx context.Context) error
	FindAppointment(ctx context.Context, params AvailableAppointmentsParams) ([]Appointment, error)
	BookAppointment(ctx context.Context, params BookAppointmentParams) (*BookedAppointment, error)
	CancelAppointment(ctx context.Context, params CancelAppointmentParams) (*CancelledAppointment, error)
	CreateOnboarding(ctx context.Context, params CreateOnboardingParams) (*Onboarding, error)
}
