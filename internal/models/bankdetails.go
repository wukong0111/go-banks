package models

// BankDetails is a common interface for bank detail responses
type BankDetails interface {
	GetBank() *Bank
	GetType() BankDetailsType
}

// BankDetailsType represents the type of bank details response
type BankDetailsType string

const (
	BankDetailsTypeSingle   BankDetailsType = "single"
	BankDetailsTypeMultiple BankDetailsType = "multiple"
)

// BankWithEnvironment represents a bank with a specific environment configuration
type BankWithEnvironment struct {
	Bank
	EnvironmentConfig *BankEnvironmentConfig `json:"environment_config"`
}

// GetBank returns the bank information
func (b *BankWithEnvironment) GetBank() *Bank {
	return &b.Bank
}

// GetType returns the type of bank details
func (b *BankWithEnvironment) GetType() BankDetailsType {
	return BankDetailsTypeSingle
}

// BankWithEnvironments represents a bank with all its environment configurations
type BankWithEnvironments struct {
	Bank
	EnvironmentConfigs map[string]*BankEnvironmentConfig `json:"environment_configs"`
}

// GetBank returns the bank information
func (b *BankWithEnvironments) GetBank() *Bank {
	return &b.Bank
}

// GetType returns the type of bank details
func (b *BankWithEnvironments) GetType() BankDetailsType {
	return BankDetailsTypeMultiple
}
