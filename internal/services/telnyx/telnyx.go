package telnyx

import (
	"github.com/ali-gulzar/speechory-core/internal"

	telnyxsdk "github.com/team-telnyx/telnyx-go/v4"
	"github.com/team-telnyx/telnyx-go/v4/option"
)

type Telnyx struct {
	Client telnyxsdk.Client
}

func New(cfg internal.TelnyxConfig) *Telnyx {
	return &Telnyx{
		Client: telnyxsdk.NewClient(option.WithAPIKey(cfg.APIKey)),
	}
}
