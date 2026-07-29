package contracts

import "time"

type FindAppointmentRequest struct {
	AssistantID string    `form:"assistant_id" binding:"required"`
	StartDate   time.Time `form:"start_date" time_format:"2006-01-02"`
}

type Appointment struct {
	Time        time.Time `json:"time"`
	DoctorsName string    `json:"doctors_name"`
	DoctorID    string    `json:"doctor_id"`
}

type FindAppointmentsResponse struct {
	Data []Appointment `json:"data"`
}

type BookAppointmentRequest struct {
	AssistantID string         `json:"assistant_id" binding:"required"`
	DoctorID    string         `json:"doctor_id" binding:"required"`
	Time        time.Time      `json:"time" binding:"required"`
	Reason      string         `json:"reason" binding:"required"`
	Patient     PatientContact `json:"patient" binding:"required"`
}

type PatientContact struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	DateOfBirth string `json:"date_of_birth" binding:"required,datetime=2006-01-02"`
}

type BookedAppointment struct {
	AppointmentID string    `json:"appointment_id"`
	Time          time.Time `json:"time"`
	ProviderID    string    `json:"provider_id"`
	Confirmed     bool      `json:"confirmed"`
}

type BookAppointmentResponse struct {
	Data BookedAppointment `json:"data"`
}

type CancelAppointmentRequest struct {
	AssistantID string          `json:"assistant_id" binding:"required"`
	Patient     PatientIdentity `json:"patient"`
}

type PatientIdentity struct {
	Name        string `json:"name" binding:"required"`
	Phone       string `json:"phone" binding:"required_without=DateOfBirth"`
	DateOfBirth string `json:"date_of_birth" binding:"required_without=Phone"`
}

type CancelledAppointment struct {
	AppointmentID string    `json:"appointment_id"`
	Time          time.Time `json:"time"`
	DoctorID      string    `json:"doctor_id"`
}

type CancelAppointmentResponse struct {
	Data CancelledAppointment `json:"data"`
}
