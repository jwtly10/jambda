package service

import (
	"testing"

	"github.com/jwtly10/jambda/api/data"
	"github.com/jwtly10/jambda/internal/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestValidateConfig(t *testing.T) {
	logger := logging.NewLogger(false, zapcore.DebugLevel)
	validator := NewConfigValidator(logger)

	tests := []struct {
		name    string
		config  *data.FunctionConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config REST",
			config: &data.FunctionConfig{
				Type:    "REST",
				Trigger: "http",
				Image:   "golang:1.22",
				Port:    new(int),
			},
			wantErr: false,
		},
		{
			name: "valid config SINGLE",
			config: &data.FunctionConfig{
				Type:    "SINGLE",
				Trigger: "cron",
				Image:   "openjdk:21-jdk",
				Port:    new(int),
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			config: &data.FunctionConfig{
				Type:    "INVALID",
				Trigger: "http",
				Image:   "golang:1.22",
			},
			wantErr: true,
			errMsg:  "invalid type 'INVALID'; must be one of 'REST' or 'SINGLE'",
		},
		{
			name: "invalid trigger",
			config: &data.FunctionConfig{
				Type:    "REST",
				Trigger: "invalid",
				Image:   "golang:1.22",
			},
			wantErr: true,
			errMsg:  "invalid trigger 'invalid'; must be one of 'http' or 'cron'",
		},
		{
			name: "invalid image",
			config: &data.FunctionConfig{
				Type:    "REST",
				Trigger: "http",
				Image:   "unknown:1.22",
			},
			wantErr: true,
			errMsg:  "invalid image 'unknown:1.22'; must be 'golang:1.22' or 'openjdk:21-jdk' or 'openjdk:17-jdk'",
		},
		{
			name: "invalid port too low",
			config: &data.FunctionConfig{
				Type:    "REST",
				Trigger: "http",
				Image:   "golang:1.22",
				Port:    new(int),
			},
			wantErr: true,
			errMsg:  "port must be between 1024 and 65535; got 0",
		},
		{
			name: "invalid port too high",
			config: &data.FunctionConfig{
				Type:    "REST",
				Trigger: "http",
				Image:   "golang:1.22",
				Port:    new(int),
			},
			wantErr: true,
			errMsg:  "port must be between 1024 and 65535; got 70000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Port != nil {
				// Set test-specific port values
				if tt.name == "invalid port too low" {
					*tt.config.Port = 0
				} else if tt.name == "invalid port too high" {
					*tt.config.Port = 70000
				} else {
					*tt.config.Port = 8080 // A valid port number
				}
			}

			err := validator.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
