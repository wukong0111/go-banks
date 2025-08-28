package models

// BankWithEnvironment represents a bank with a specific environment configuration
type BankWithEnvironment struct {
	Bank
	EnvironmentConfig *BankEnvironmentConfig `json:"environment_config"`
}

// BankWithEnvironments represents a bank with all its environment configurations
type BankWithEnvironments struct {
	Bank
	EnvironmentConfigs map[string]*BankEnvironmentConfig `json:"environment_configs"`
}
