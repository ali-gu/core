package ehr

import "time"

type AvailableAppointmentsParams struct {
	Subdomain  string
	LocationID string
	StartDate  time.Time
	Days       int
}

type Appointment struct {
	Time         time.Time
	ProviderID   string
	ProviderName string
}

type BookAppointmentParams struct {
	Subdomain  string
	LocationID string
	ProviderID string
	Time       time.Time
	Note       string
	Patient    PatientDetails
}

type PatientDetails struct {
	FirstName   string
	LastName    string
	Phone       string
	Email       string
	DateOfBirth string
}

type BookedAppointment struct {
	ID         string
	Time       time.Time
	ProviderID string
	Confirmed  bool
}

type CancelAppointmentParams struct {
	Subdomain  string
	LocationID string
	Patient    PatientLookup
}

type PatientLookup struct {
	Name        string
	Phone       string
	DateOfBirth string
}

type CancelledAppointment struct {
	ID         string
	Time       time.Time
	ProviderID string
}

type CreateOnboardingParams struct {
	InstitutionName    string
	InstitutionEmail   string
	InstitutionZipCode string
	InstitutionWebsite string
	Subdomain          string
	EHRName            string
}

type Onboarding struct {
	ID           string
	Subdomain    string
	LocationID   string
	URL          string
	URLExpiresAt time.Time
	Status       string
}
