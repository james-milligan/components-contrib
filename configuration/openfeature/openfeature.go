package openfeature

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dapr/components-contrib/configuration"
	internal "github.com/dapr/components-contrib/configuration/openfeature/pkg"
	"github.com/dapr/kit/logger"
	"github.com/open-feature/go-sdk/pkg/openfeature"
)

type FlagType string
type ProviderType string

const (
	keyProvider                        = "provider"
	ProviderFlagd         ProviderType = "flagd"
	ProviderGoFeatureFlag ProviderType = "goFeatureFlag"

	keyDefaultValueSuffix          = "_defaultValue"
	keyTypeSuffix                  = "_type"
	FlagTypeBool          FlagType = "bool"
	FlagTypeString        FlagType = "string"
	FlagTypeInt           FlagType = "int"
	FlagTypeFloat         FlagType = "float"
	FlagTypeObject        FlagType = "object"
)

type ConfigurationStore struct {
	client openfeature.Client
}

func NewOpenFeatureConfigurationStore(logger logger.Logger) configuration.Store {
	return &ConfigurationStore{}
}

func (r *ConfigurationStore) Init(ctx context.Context, metadata configuration.Metadata) error {
	provider, ok := metadata.Properties[keyProvider]
	if !ok {
		return fmt.Errorf("no provider defined in metadata")
	}
	switch ProviderType(provider) {
	case ProviderFlagd:
		p, err := internal.NewFlagdProvider(metadata)
		if err != nil {
			return err
		}
		openfeature.SetProvider(p)
	case ProviderGoFeatureFlag:
		p, err := internal.NewGoFeatureFlagProvider(ctx, metadata)
		if err != nil {
			return err
		}
		openfeature.SetProvider(p)
	default:
		return fmt.Errorf("unrecognized provider %s", provider)
	}
	r.client = *openfeature.NewClient("dapr")
	return nil
}

func (r *ConfigurationStore) Get(ctx context.Context, req *configuration.GetRequest) (*configuration.GetResponse, error) {
	res := &configuration.GetResponse{}
	attributes := map[string]interface{}{}
	for k, v := range req.Metadata {
		attributes[k] = v
	}
	evalContext := openfeature.NewEvaluationContext(
		"",
		attributes,
	)
	for _, key := range req.Keys {
		var item *configuration.Item
		var err error
		typeKey := fmt.Sprintf("%s%s", key, keyTypeSuffix)
		t, ok := req.Metadata[typeKey]
		if !ok {
			return nil, fmt.Errorf("required key %s not found in metadata for flag %s", typeKey, key)
		}
		defaultValueKey := fmt.Sprintf("%s%s", key, keyDefaultValueSuffix)
		defaultValueString, ok := req.Metadata[defaultValueKey]
		if !ok {
			return nil, fmt.Errorf("required key %s not found in metadata for flag %s", defaultValueKey, key)
		}
		switch FlagType(t) {
		case FlagTypeBool:
			item, err = r.getBoolFlag(ctx, key, evalContext, defaultValueString)
		case FlagTypeString:
			item, err = r.getStringFlag(ctx, key, evalContext, defaultValueString)
		case FlagTypeInt:
			item, err = r.getIntFlag(ctx, key, evalContext, defaultValueString)
		case FlagTypeFloat:
			item, err = r.getFloatFlag(ctx, key, evalContext, defaultValueString)
		case FlagTypeObject:
			item, err = r.getObjectFlag(ctx, key, evalContext, defaultValueString)
		}
		if err != nil {
			return nil, err
		}
		res.Items[key] = item
	}
	return res, nil
}

func (r *ConfigurationStore) GetComponentMetadata() map[string]string {
	metadata := r.client.Metadata()
	res := map[string]string{}
	res["name"] = metadata.Name()
	return res
}

func (r *ConfigurationStore) Subscribe(ctx context.Context, req *configuration.SubscribeRequest, handler configuration.UpdateHandler) (string, error) {
	return "", nil
}

func (r *ConfigurationStore) Unsubscribe(ctx context.Context, req *configuration.UnsubscribeRequest) error {
	return nil
}

func (r *ConfigurationStore) getBoolFlag(ctx context.Context, flag string, evalContext openfeature.EvaluationContext, d string) (*configuration.Item, error) {
	defaultValue, err := strconv.ParseBool(d)
	if err != nil {
		return nil, err
	}
	res, err := r.client.BooleanValueDetails(ctx, flag, defaultValue, evalContext)
	if err != nil {
		return nil, err
	}
	return &configuration.Item{
		Value: fmt.Sprintf("%t", res.Value),
		Metadata: map[string]string{
			"flagKey": res.FlagKey,
			"variant": res.Variant,
			"reason":  string(res.Reason),
		},
	}, nil
}

func (r *ConfigurationStore) getStringFlag(ctx context.Context, flag string, evalContext openfeature.EvaluationContext, defaultValue string) (*configuration.Item, error) {
	res, err := r.client.StringValueDetails(ctx, flag, defaultValue, evalContext)
	if err != nil {
		return nil, err
	}
	return &configuration.Item{
		Value: res.Value,
		Metadata: map[string]string{
			"flagKey": res.FlagKey,
			"variant": res.Variant,
			"reason":  string(res.Reason),
		},
	}, nil
}

func (r *ConfigurationStore) getIntFlag(ctx context.Context, flag string, evalContext openfeature.EvaluationContext, d string) (*configuration.Item, error) {
	defaultValue, err := strconv.ParseInt(d, 10, 64)
	if err != nil {
		return nil, err
	}
	res, err := r.client.IntValueDetails(ctx, flag, int64(defaultValue), evalContext)
	if err != nil {
		return nil, err
	}
	return &configuration.Item{
		Value: fmt.Sprintf("%d", res.Value),
		Metadata: map[string]string{
			"flagKey": res.FlagKey,
			"variant": res.Variant,
			"reason":  string(res.Reason),
		},
	}, nil
}

func (r *ConfigurationStore) getFloatFlag(ctx context.Context, flag string, evalContext openfeature.EvaluationContext, d string) (*configuration.Item, error) {
	defaultValue, err := strconv.ParseFloat(d, 64)
	if err != nil {
		return nil, err
	}
	res, err := r.client.FloatValueDetails(ctx, flag, defaultValue, evalContext)
	if err != nil {
		return nil, err
	}
	return &configuration.Item{
		Value: fmt.Sprintf("%f", res.Value),
		Metadata: map[string]string{
			"flagKey": res.FlagKey,
			"variant": res.Variant,
			"reason":  string(res.Reason),
		},
	}, nil
}

func (r *ConfigurationStore) getObjectFlag(ctx context.Context, flag string, evalContext openfeature.EvaluationContext, d string) (*configuration.Item, error) {
	defaultValue := map[string]interface{}{}
	err := json.Unmarshal([]byte(d), &defaultValue)
	if err != nil {
		return nil, err
	}
	res, err := r.client.ObjectValueDetails(ctx, flag, defaultValue, evalContext)
	if err != nil {
		return nil, err
	}
	val, err := json.Marshal(res.Value)
	if err != nil {
		return nil, err
	}
	return &configuration.Item{
		Value: string(val),
		Metadata: map[string]string{
			"flagKey": res.FlagKey,
			"variant": res.Variant,
			"reason":  string(res.Reason),
		},
	}, nil
}
