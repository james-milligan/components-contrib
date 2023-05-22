package pkg

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dapr/components-contrib/configuration"
	flagd "github.com/open-feature/go-sdk-contrib/providers/flagd/pkg"
	gofeatureflag "github.com/open-feature/go-sdk-contrib/providers/go-feature-flag/pkg"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

const (
	keyFlagdHost = "flagdHost"
	keyFlagdPort = "flagdPort"

	keyGoFeatureFlagEndpoint = "goFeatureFlagEndpoint"
	keyGoFeatureFlagTimeout  = "goFeatureFlagTimeout"
)

func NewFlagdProvider(metadata configuration.Metadata) (openfeature.FeatureProvider, error) {
	args := []flagd.ProviderOption{}
	host, ok := metadata.Properties[keyFlagdHost]
	if ok {
		args = append(args, flagd.WithHost(host))
	}
	port, ok := metadata.Properties[keyFlagdPort]
	if ok {
		p, err := strconv.Atoi(port)
		if err != nil {
			return nil, fmt.Errorf("could not parse metadata value %s, %s", keyFlagdPort, port)
		}
		args = append(args, flagd.WithPort(uint16(p)))
	}
	return flagd.NewProvider(args...), nil
}

func NewGoFeatureFlagProvider(ctx context.Context, metadata configuration.Metadata) (*gofeatureflag.Provider, error) {
	endpoint, ok := metadata.Properties[keyGoFeatureFlagEndpoint]
	if !ok {
		return nil, fmt.Errorf("missing required metadata value %s", keyGoFeatureFlagEndpoint)
	}
	timeout := 1 * time.Second
	t, ok := metadata.Properties[keyFlagdPort]
	if ok {
		tt, err := strconv.Atoi(t)
		if err != nil {
			return nil, fmt.Errorf("could not parse metadata value %s, %s", keyFlagdPort, t)
		}
		timeout = time.Duration(tt) * time.Second
	}
	options := gofeatureflag.ProviderOptions{
		Endpoint: endpoint,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
	provider, err := gofeatureflag.NewProviderWithContext(ctx, options)
	if err != nil {
		return nil, err
	}
	return provider, nil
}
