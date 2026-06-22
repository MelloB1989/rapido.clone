package config

import (
	c "github.com/MelloB1989/karma/config"
)

// Config holds runtime configuration sourced from environment variables.
// On Lambda these are injected by the CDK stack (and AWS_REGION by the runtime);
// locally they fall back to sensible defaults.
type Config struct {
	Port string

	// Region of the Cognito User Pool. On Lambda this comes from the reserved
	// AWS_REGION variable; locally it falls back to COGNITO_REGION / a default.
	Region string

	// CognitoUserPoolID is the pool whose tokens we trust.
	CognitoUserPoolID string

	// CognitoClientID, when set, is checked against the token's client_id claim.
	// Leave empty to skip audience/client validation.
	CognitoClientID string

	// TableName is the DynamoDB single-table name. Injected by the CDK stack on
	// Lambda; defaults to "raftaar" locally.
	UsersTableName string
}

// Load reads configuration from the environment.
func Load() Config {
	region := c.GetEnvOrDefault("AWS_REGION", c.GetEnvRaw("COGNITO_REGION"))
	if region == "" {
		region = "ap-south-1"
	}

	return Config{
		Port:              c.GetEnvOrDefault("PORT", "8080"),
		Region:            region,
		CognitoUserPoolID: c.GetEnvRaw("COGNITO_USER_POOL_ID"),
		CognitoClientID:   c.GetEnvRaw("COGNITO_CLIENT_ID"),
		UsersTableName:    c.GetEnvOrDefault("DYNAMODB_TABLE", "raftaar-users"),
	}
}
