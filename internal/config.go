package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	cfgfiles "github.com/ali-gulzar/speechory-core/config"
	"github.com/ali-gulzar/speechory-core/internal/constants"
)

type DBEndpointConfig struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}

type SSLMode string

const (
	SSLModeDisabled SSLMode = "disabled"
	SSLModeRequired SSLMode = "required"
)

type DBConfig struct {
	Writer   DBEndpointConfig `json:"writer"`
	Reader   DBEndpointConfig `json:"reader"`
	DBName   string           `json:"db_name"`
	Username string           `json:"username"`
	Password string           `json:"password"`
	SSLMode  SSLMode          `json:"ssl_mode"`
}

type TelnyxConfig struct {
	APIKey    string `json:"api_key"`
	PublicKey string `json:"public_key"`
}

type NexHealthConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

type SupabaseConfig struct {
	URL    string `json:"url"`
	PubKey string `json:"pub_key"`
}

type Config struct {
	Landscape constants.Landscape `json:"landscape"`
	DB        DBConfig            `json:"db"`
	Domain    string              `json:"domain"`
	Telnyx    TelnyxConfig        `json:"telnyx"`
	NexHealth NexHealthConfig     `json:"nexhealth"`
	Supabase  SupabaseConfig      `json:"supabase"`
}

func NewConfig(landscape constants.Landscape) (*Config, error) {
	var config Config
	config.Landscape = landscape

	baseFile := fmt.Sprintf("%s.json", landscape)

	cfgFile, err := cfgfiles.Files.ReadFile(baseFile)
	if err != nil {
		return nil, err
	}

	cfgFile, err = renderConfigTemplate(baseFile, cfgFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(cfgFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func renderConfigTemplate(name string, raw []byte) ([]byte, error) {
	tmpl, err := template.New(name).Funcs(template.FuncMap{"env": envLookup}).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing config template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return nil, fmt.Errorf("rendering config template: %w", err)
	}

	return buf.Bytes(), nil
}

func envLookup(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("config references environment variable %q, which is not set", key)
	}
	return val, nil
}
