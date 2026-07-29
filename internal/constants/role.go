package constants

type RoleType string

const (
	RoleTypeSuperAdmin RoleType = "SUPER_ADMIN"
	RoleTypeAdmin      RoleType = "ADMIN"
	RoleTypeReader     RoleType = "READER"
)

func (r RoleType) IsValid() bool {
	switch r {
	case RoleTypeSuperAdmin, RoleTypeAdmin, RoleTypeReader:
		return true
	default:
		return false
	}
}
