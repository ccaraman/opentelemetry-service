// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configtls

import (
	"crypto/x509"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsToConfig(t *testing.T) {
	tests := []struct {
		name        string
		options     TLSConfig
		fakeSysPool bool
		expectError string
	}{
		{
			name:    "should load system CA",
			options: TLSConfig{CAFile: ""},
		},
		{
			name:        "should fail with fake system CA",
			fakeSysPool: true,
			options:     TLSConfig{CAFile: ""},
			expectError: "fake system pool",
		},
		{
			name:    "should load custom CA",
			options: TLSConfig{CAFile: "testdata/testCA.pem"},
		},
		{
			name:        "should fail with invalid CA file path",
			options:     TLSConfig{CAFile: "testdata/not/valid"},
			expectError: "failed to load CA",
		},
		{
			name:        "should fail with invalid CA file content",
			options:     TLSConfig{CAFile: "testdata/testCA-bad.txt"},
			expectError: "failed to parse CA",
		},
		{
			name: "should load valid TLS  settings",
			options: TLSConfig{
				CAFile:   "testdata/testCA.pem",
				CertFile: "testdata/test-cert.pem",
				KeyFile:  "testdata/test-key.pem",
			},
		},
		{
			name: "should fail with missing TLS KeyFile",
			options: TLSConfig{
				CAFile:   "testdata/testCA.pem",
				CertFile: "testdata/test-cert.pem",
			},
			expectError: "both certificate and key must be supplied",
		},
		{
			name: "should fail with invalid TLS KeyFile",
			options: TLSConfig{
				CAFile:   "testdata/testCA.pem",
				CertFile: "testdata/test-cert.pem",
				KeyFile:  "testdata/not/valid",
			},
			expectError: "failed to load TLS cert and key",
		},
		{
			name: "should fail with missing TLS Cert",
			options: TLSConfig{
				CAFile:  "testdata/testCA.pem",
				KeyFile: "testdata/test-key.pem",
			},
			expectError: "both certificate and key must be supplied",
		},
		{
			name: "should fail with invalid TLS Cert",
			options: TLSConfig{
				CAFile:   "testdata/testCA.pem",
				CertFile: "testdata/not/valid",
				KeyFile:  "testdata/test-key.pem",
			},
			expectError: "failed to load TLS cert and key",
		},
		{
			name: "should fail with invalid TLS CA",
			options: TLSConfig{
				CAFile: "testdata/not/valid",
			},
			expectError: "failed to load CA",
		},
		{
			name: "should fail with invalid CA pool",
			options: TLSConfig{
				CAFile: "testdata/testCA-bad.txt",
			},
			expectError: "failed to parse CA",
		},
		{
			name: "should pass with valid CA pool",
			options: TLSConfig{
				CAFile: "testdata/testCA.pem",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.fakeSysPool {
				saveSystemCertPool := systemCertPool
				systemCertPool = func() (*x509.CertPool, error) {
					return nil, fmt.Errorf("fake system pool")
				}
				defer func() {
					systemCertPool = saveSystemCertPool
				}()
			}
			cfg, err := test.options.LoadTLSConfig()
			if test.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.expectError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, cfg)
			}
		})
	}
}
