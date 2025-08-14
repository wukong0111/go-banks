package models

import (
	"time"
)

type EnvironmentType string

const (
	EnvironmentSandbox    EnvironmentType = "sandbox"
	EnvironmentProduction EnvironmentType = "production"
	EnvironmentUAT        EnvironmentType = "uat"
	EnvironmentTest       EnvironmentType = "test"
)

type BankEnvironmentConfig struct {
	BankID                       string          `json:"bank_id" db:"bank_id"`
	Environment                  EnvironmentType `json:"environment" db:"environment"`
	Enabled                      bool            `json:"enabled" db:"enabled"`
	Blocked                      bool            `json:"blocked" db:"blocked"`
	BlockedText                  *string         `json:"blocked_text" db:"blocked_text"`
	Risky                        bool            `json:"risky" db:"risky"`
	RiskyMessage                 *string         `json:"risky_message" db:"risky_message"`
	SupportsInstantPayments      *bool           `json:"supports_instant_payments" db:"supports_instant_payments"`
	InstantPaymentsActivated     *bool           `json:"instant_payments_activated" db:"instant_payments_activated"`
	InstantPaymentsLimit         *int32          `json:"instant_payments_limit" db:"instant_payments_limit"`
	OkStatusCodesSimplePayment   map[string]any  `json:"ok_status_codes_simple_payment" db:"ok_status_codes_simple_payment"`
	OkStatusCodesInstantPayment  map[string]any  `json:"ok_status_codes_instant_payment" db:"ok_status_codes_instant_payment"`
	OkStatusCodesPeriodicPayment map[string]any  `json:"ok_status_codes_periodic_payment" db:"ok_status_codes_periodic_payment"`
	EnabledPeriodicPayment       *bool           `json:"enabled_periodic_payment" db:"enabled_periodic_payment"`
	FrequencyPeriodicPayment     *string         `json:"frequency_periodic_payment" db:"frequency_periodic_payment"`
	ConfigPeriodicPayment        *string         `json:"config_periodic_payment" db:"config_periodic_payment"`
	AppAuthSetupRequired         bool            `json:"app_auth_setup_required" db:"app_auth_setup_required"`
	CreatedAt                    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time       `json:"updated_at" db:"updated_at"`
}
