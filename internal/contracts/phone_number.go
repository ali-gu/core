package contracts

import (
	"time"

	"github.com/segmentio/ksuid"
)

type PurchasedPhoneNumber struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	Status      string `json:"status"`
}

type ListPurchasedPhoneNumbersResponse struct {
	Data []PurchasedPhoneNumber `json:"data"`
}

type ListAvailablePhoneNumbersRequest struct {
	CountryCode string `form:"country_code"`
	AreaCode    string `form:"area_code"`
	Contains    string `form:"contains"`
}

type AvailablePhoneNumber struct {
	PhoneNumber string `json:"phone_number"`
	Reservable  bool   `json:"reservable"`
}

type ListAvailablePhoneNumbersResponse struct {
	Data []AvailablePhoneNumber `json:"data"`
}

type ReservePhoneNumberRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type ReservePhoneNumberResponse struct {
	ID          ksuid.KSUID `json:"id"`
	PhoneNumber string      `json:"phone_number"`
	EntityState string      `json:"entity_state"`
	CreatedAt   time.Time   `json:"created_at"`
}

type PhoneNumberURI struct {
	PhoneNumberID ksuid.KSUID `uri:"phone_number_id,parser=encoding.TextUnmarshaler"`
}

type DisablePhoneNumberResponse struct {
	ID          ksuid.KSUID `json:"id"`
	PhoneNumber string      `json:"phone_number"`
	EntityState string      `json:"entity_state"`
	DisabledAt  *time.Time  `json:"disabled_at"`
}

type ActivatePhoneNumberRequest struct {
	PhoneNumberID  ksuid.KSUID `json:"phone_number_id" binding:"required"`
	PhoneNumberRef string      `json:"phone_number_ref" binding:"required"`
}

type ActivatePhoneNumberResponse struct {
	ID               ksuid.KSUID `json:"id"`
	PhoneNumber      string      `json:"phone_number"`
	PhoneNumberIDRef *string     `json:"phone_number_id_ref"`
	EntityState      string      `json:"entity_state"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        *time.Time  `json:"updated_at"`
}

type AdminCreatePhoneNumberRequest struct {
	PhoneNumber      string `json:"phone_number" binding:"required"`
	PhoneNumberIDRef string `json:"phone_number_id_ref" binding:"required"`
}

type PhoneNumber struct {
	ID          ksuid.KSUID `json:"id"`
	PhoneNumber string      `json:"phone_number"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
}

type GetPhoneNumbersResponse struct {
	Data []PhoneNumber `json:"data"`
}
