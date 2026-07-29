package supabase

import (
	"github.com/ali-gulzar/speechory-core/internal"
	supa "github.com/supabase-community/supabase-go"
)

type Supabase struct {
	Client *supa.Client
}

func New(config internal.SupabaseConfig) *Supabase {
	client, err := supa.NewClient(config.URL, config.PubKey, &supa.ClientOptions{})
	if err != nil {
		panic(err)
	}
	return &Supabase{
		Client: client,
	}
}
