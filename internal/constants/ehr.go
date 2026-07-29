package constants

type EHR string

const (
	EHRNexHealth EHR = "NEXHEALTH"
)

const DefaultAppointmentSearchDays = 7

func (e EHR) IsValid() bool {
	return e == EHRNexHealth
}
