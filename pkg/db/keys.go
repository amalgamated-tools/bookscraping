package db

import "context"

const (
	// Config keys for Booklore credentials and tokens
	BookloreServerURL = "booklore_server_url"
	BookloreUsername  = "booklore_username"
	BooklorePassword  = "booklore_password"
	BookloreToken     = "booklore_access_token"
	BookloreRefToken  = "booklore_refresh_token"
)

func GetAllConfigKeys() []string {
	return []string{
		BookloreServerURL,
		BookloreUsername,
		BooklorePassword,
		BookloreToken,
		BookloreRefToken,
	}
}

func GetAllConfig(ctx context.Context, q Querier) (map[string]string, error) {
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
