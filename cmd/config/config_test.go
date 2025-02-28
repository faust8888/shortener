package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestInitializeConfig(t *testing.T) {
	tests := []struct {
		name                 string
		environmentVariables map[string]string
		flagVariables        map[string]string
	}{
		{
			name: "Successfully used default flag values",
		},
		{
			name:                 "Using environment variables with priority over flags",
			environmentVariables: map[string]string{ServerAddressEnvironmentName: "localhost:7777"},
		},
		{
			name:          "Using custom flag values",
			flagVariables: map[string]string{ServerAddressFlag: "localhost:3333", BaseShortURLFlag: "http://localhost:3333"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnvironmentVariables(t, tt.environmentVariables)
			LoadConfig()
			setFlagVariables(tt.flagVariables)

			if tt.environmentVariables == nil && tt.flagVariables == nil {
				t.Log("Validating flag default values")
				assert.NotEmpty(t, Cfg.ServerAddress)
				assert.NotEmpty(t, Cfg.BaseShortURL)
			}
			if tt.environmentVariables != nil && tt.flagVariables == nil {
				t.Log("Validating priority of environment variables")
				for envName, envValue := range tt.environmentVariables {
					if envName == ServerAddressEnvironmentName {
						assert.Equal(t, envValue, Cfg.ServerAddress)
					}
					if envName == BaseShortURLEnvironmentName {
						assert.Equal(t, envValue, Cfg.BaseShortURL)
					}
				}
			}
			if tt.flagVariables != nil && tt.environmentVariables == nil {
				t.Log("Validating flag custom values")
				for flagName, flagValue := range tt.flagVariables {
					if flagName == ServerAddressFlag {
						assert.Equal(t, flagValue, Cfg.ServerAddress)
					}
					if flagName == BaseShortURLFlag {
						assert.Equal(t, flagValue, Cfg.BaseShortURL)
					}
				}
			}
		})
	}
}

func setEnvironmentVariables(t *testing.T, environmentVariables map[string]string) {
	for envName, envValue := range environmentVariables {
		err := os.Setenv(envName, envValue)
		require.NoError(t, err)
	}
}

func setFlagVariables(flags map[string]string) {
	if len(flags) != 0 {
		os.Args = append([]string{}, []string{""}...)
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		for flagName, flagValue := range flags {
			if flagName == ServerAddressFlag {
				flag.StringVar(&Cfg.ServerAddress, flagName, flagValue, "Server address ")
			}
			if flagName == BaseShortURLFlag {
				flag.StringVar(&Cfg.BaseShortURL, flagName, flagValue, "Base short URL")
			}
		}
		flag.Parse()
	}
}
