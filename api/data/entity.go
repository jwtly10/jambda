package data

import "time"

type FunctionEntity struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	ExternalId    string          `json:"external_id"`
	State         string          `json:"state"`
	Configuration *FunctionConfig `json:"configuration,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type FunctionConfig struct {
	Trigger string            `json:"trigger"`
	Image   string            `json:"image"`
	Type    string            `json:"type"`
	Port    *int              `json:"port,omitempty"`
	EnvVars map[string]string `json:"env_vars,omitempty"`
}
