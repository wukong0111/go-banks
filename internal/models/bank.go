package models

import (
	"time"

	"github.com/google/uuid"
)

type Bank struct {
	BankID                 string                 `json:"bank_id" db:"bank_id"`
	Name                   string                 `json:"name" db:"name"`
	BankCodes              []string               `json:"bank_codes" db:"bank_codes"`
	BIC                    *string                `json:"bic" db:"bic"`
	RealName               *string                `json:"real_name" db:"real_name"`
	API                    string                 `json:"api" db:"api"`
	APIVersion             string                 `json:"api_version" db:"api_version"`
	ASPSP                  string                 `json:"aspsp" db:"aspsp"`
	ProductCode            *string                `json:"product_code" db:"product_code"`
	Country                string                 `json:"country" db:"country"`
	BankGroupID            *uuid.UUID             `json:"bank_group_id" db:"bank_group_id"`
	LogoURL                *string                `json:"logo_url" db:"logo_url"`
	Documentation          *string                `json:"documentation" db:"documentation"`
	Keywords               map[string]any `json:"keywords" db:"keywords"`
	Attribute              map[string]any `json:"attribute" db:"attribute"`
	AuthTypeChoiceRequired bool                   `json:"auth_type_choice_required" db:"auth_type_choice_required"`
	CreatedAt              time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at" db:"updated_at"`
}
