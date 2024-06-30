package routes

import (
	"net/url"
	"testing"
)

func TestCanParseUrlFromContainerBase(t *testing.T) {
	tests := []struct {
		name        string
		baseUrl     string
		proxiedUrl  string
		expectedUrl string
		expectError bool
	}{
		{
			name:        "Valid URL",
			baseUrl:     "http://localhost:8000",
			proxiedUrl:  "/v1/api/function/57d4b724/getUser",
			expectedUrl: "http://localhost:8000/getUser",
			expectError: false,
		},
		{
			name:        "Valid URL with query",
			baseUrl:     "http://localhost:8000",
			proxiedUrl:  "/v1/api/function/57d4b724/getUser?query=param",
			expectedUrl: "http://localhost:8000/getUser?query=param",
			expectError: false,
		},
		{
			name:        "Valid URL with no sub path",
			baseUrl:     "http://localhost:8000",
			proxiedUrl:  "/v1/api/function/57d4b724",
			expectedUrl: "http://localhost:8000/",
			expectError: false,
		},
		{
			name:        "Valid URL with no sub path and query",
			baseUrl:     "http://localhost:8000",
			proxiedUrl:  "/v1/api/function/57d4b724?id=1",
			expectedUrl: "http://localhost:8000/?id=1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxiedUrl, _ := url.Parse(tt.proxiedUrl)
			resultUrl, err := parseProxiedUrlGivenBaseUrl(tt.baseUrl, proxiedUrl)
			if (err != nil) != tt.expectError {
				t.Errorf("parseProxiedUrlGivenBaseUrl() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if !tt.expectError && resultUrl.String() != tt.expectedUrl {
				t.Errorf("parseProxiedUrlGivenBaseUrl() = %v, want %v", resultUrl.String(), tt.expectedUrl)
			}
		})
	}

}
