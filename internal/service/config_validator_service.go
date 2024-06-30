package service

import (
	"fmt"

	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/logging"
)

type ConfigValidator struct {
	log logging.Logger
}

func NewConfigValidator(log logging.Logger) *ConfigValidator {
	return &ConfigValidator{
		log: log,
	}
}

func (cv *ConfigValidator) ValidateConfig(config *data.FunctionConfig) error {
	// Validate Type
	validTypes := map[string]bool{"REST": true, "SINGLE": true}
	if _, ok := validTypes[config.Type]; !ok {
		return fmt.Errorf("invalid type '%s'; must be one of 'REST' or 'SINGLE'", config.Type)
	}

	// Validate Trigger
	validTriggers := map[string]bool{"http": true, "cron": true}
	if _, ok := validTriggers[config.Trigger]; !ok {
		return fmt.Errorf("invalid trigger '%s'; must be one of 'http' or 'cron'", config.Trigger)
	}

	// Validate Image (example: must not be empty)
	validImages := map[string]bool{"golang:1.22": true}
	if _, ok := validImages[config.Image]; !ok {
		return fmt.Errorf("invalid image '%s'; must be 'golang:1.22'", config.Trigger)
	}

	// Optional: Validate Port
	if config.Port != nil && (*config.Port < 1024 || *config.Port > 65535) {
		return fmt.Errorf("port must be between 1024 and 65535; got %d", *config.Port)
	}

	// TODO: Validate env vars

	return nil
}
