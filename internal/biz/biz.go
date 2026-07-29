package biz

import (
	"github.com/ali-gulzar/speechory-core/internal/services/agent"
	"github.com/ali-gulzar/speechory-core/internal/services/auth"
	"github.com/ali-gulzar/speechory-core/internal/services/ehr"
	"github.com/ali-gulzar/speechory-core/internal/services/phonenumber"
	"github.com/ali-gulzar/speechory-core/internal/services/tool"
	"github.com/ali-gulzar/speechory-core/internal/storage"
)

type Biz struct {
	Common       ICommon
	Agent        IAgent
	Practice     IPractice
	Location     ILocation
	PhoneNumber  IPhoneNumber
	Appointment  IAppointment
	EHR          IEHR
	Role         IRole
	User         IUser
	Analytics    IAnalytics
	Tool         ITool
	Conversation IConversation
}

type Dependencies struct {
	Storage                  storage.Storage
	TelnyxAgent              agent.IAgent
	TelnyxPhoneNumberManager phonenumber.IPhoneNumberManager
	SupabaseAuth             auth.IAuth
	EHR                      ehr.IEHR
	TelnyxTool               tool.ITool
	Domain                   string
}

func NewBiz(deps Dependencies) *Biz {
	bz := Biz{}
	bz.Common = &Common{
		Biz: &bz,
	}
	bz.Agent = &Agent{
		Biz:         &bz,
		storage:     deps.Storage,
		telnyxAgent: deps.TelnyxAgent,
	}
	bz.Practice = &Practice{
		Biz:     &bz,
		storage: deps.Storage,
	}
	bz.Location = &Location{
		Biz:     &bz,
		storage: deps.Storage,
	}
	bz.PhoneNumber = &PhoneNumber{
		Biz:                      &bz,
		storage:                  deps.Storage,
		telnyxPhoneNumberManager: deps.TelnyxPhoneNumberManager,
	}
	bz.Appointment = &Appointment{
		Biz:     &bz,
		storage: deps.Storage,
		ehr:     deps.EHR,
	}
	bz.EHR = &EHR{
		Biz:     &bz,
		storage: deps.Storage,
		ehr:     deps.EHR,
	}
	bz.Role = &Role{
		Biz:     &bz,
		storage: deps.Storage,
	}
	bz.User = &User{
		Biz:          &bz,
		storage:      deps.Storage,
		supabaseAuth: deps.SupabaseAuth,
	}
	bz.Analytics = &Analytics{
		Biz:         &bz,
		storage:     deps.Storage,
		telnyxAgent: deps.TelnyxAgent,
	}
	bz.Tool = &Tool{
		Biz:        &bz,
		storage:    deps.Storage,
		telnyxTool: deps.TelnyxTool,
		domain:     deps.Domain,
	}
	bz.Conversation = &Conversation{
		Biz:         &bz,
		storage:     deps.Storage,
		telnyxAgent: deps.TelnyxAgent,
	}
	return &bz
}
