package db

import "context"

const (
	// Config keys for Booklore credentials and tokens
	ConfigKeyServerURL        = "serverUrl"
	ConfigKeyUsername         = "username"
	ConfigKeyPassword         = "password"
	ConfigKeyBookloreToken    = "booklore_access_token"
	ConfigKeyBookloreRefToken = "booklore_refresh_token"
)

func GetAllConfigKeys() []string {
	return []string{
		ConfigKeyServerURL,
		ConfigKeyUsername,
		ConfigKeyPassword,
		ConfigKeyBookloreToken,
		ConfigKeyBookloreRefToken,
	}
}

func (q *Queries) GetAllConfig(ctx context.Context) (map[string]string, error) {
	configs, err := q.GetMultipleConfig(ctx, GetAllConfigKeys())
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, config := range configs {
		result[config.Key] = config.Value
	}
	return result, nil
}
