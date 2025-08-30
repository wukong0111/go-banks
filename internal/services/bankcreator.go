package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

type CreateBankRequest struct {
	BankID                 string                        `json:"bank_id" binding:"required"`
	Name                   string                        `json:"name" binding:"required"`
	BankCodes              []string                      `json:"bank_codes" binding:"required"`
	API                    string                        `json:"api" binding:"required"`
	APIVersion             string                        `json:"api_version" binding:"required"`
	ASPSP                  string                        `json:"aspsp" binding:"required"`
	Country                string                        `json:"country" binding:"required"`
	AuthTypeChoiceRequired bool                          `json:"auth_type_choice_required"`
	BIC                    *string                       `json:"bic,omitempty"`
	RealName               *string                       `json:"real_name,omitempty"`
	ProductCode            *string                       `json:"product_code,omitempty"`
	BankGroupID            *string                       `json:"bank_group_id,omitempty"`
	LogoURL                *string                       `json:"logo_url,omitempty"`
	Documentation          *string                       `json:"documentation,omitempty"`
	Keywords               map[string]any                `json:"keywords,omitempty"`
	Attribute              map[string]any                `json:"attribute,omitempty"`
	Environments           []string                      `json:"environments,omitempty"`
	Configuration          *EnvironmentConfig            `json:"configuration,omitempty"`
	Configurations         map[string]*EnvironmentConfig `json:"configurations,omitempty"`
}

type EnvironmentConfig struct {
	Enabled                      *bool    `json:"enabled,omitempty"`
	Blocked                      *bool    `json:"blocked,omitempty"`
	BlockedText                  *string  `json:"blocked_text,omitempty"`
	Risky                        *bool    `json:"risky,omitempty"`
	RiskyMessage                 *string  `json:"risky_message,omitempty"`
	SupportsInstantPayments      *bool    `json:"supports_instant_payments,omitempty"`
	InstantPaymentsActivated     *bool    `json:"instant_payments_activated,omitempty"`
	InstantPaymentsLimit         *int32   `json:"instant_payments_limit,omitempty"`
	OkStatusCodesSimplePayment   []string `json:"ok_status_codes_simple_payment,omitempty"`
	OkStatusCodesInstantPayment  []string `json:"ok_status_codes_instant_payment,omitempty"`
	OkStatusCodesPeriodicPayment []string `json:"ok_status_codes_periodic_payment,omitempty"`
	EnabledPeriodicPayment       *bool    `json:"enabled_periodic_payment,omitempty"`
	FrequencyPeriodicPayment     *string  `json:"frequency_periodic_payment,omitempty"`
	ConfigPeriodicPayment        *string  `json:"config_periodic_payment,omitempty"`
	AppAuthSetupRequired         *bool    `json:"app_auth_setup_required,omitempty"`
}

type BankCreator interface {
	CreateBank(ctx context.Context, request *CreateBankRequest) (*models.Bank, error)
}

type BankCreatorService struct {
	writer repository.BankWriter
}

func NewBankCreatorService(writer repository.BankWriter) *BankCreatorService {
	return &BankCreatorService{
		writer: writer,
	}
}

func (s *BankCreatorService) CreateBank(ctx context.Context, request *CreateBankRequest) (*models.Bank, error) {
	bank := s.requestToBank(request)

	switch {
	case request.Environments != nil || request.Configuration != nil:
		configs := s.buildEnvironmentConfigs(request, bank.BankID)
		if err := s.writer.CreateBankWithEnvironments(ctx, bank, configs); err != nil {
			return nil, fmt.Errorf("failed to create bank with environments: %w", err)
		}
	case request.Configurations != nil:
		configs := s.buildConfigurationsConfigs(request, bank.BankID)
		if err := s.writer.CreateBankWithEnvironments(ctx, bank, configs); err != nil {
			return nil, fmt.Errorf("failed to create bank with configurations: %w", err)
		}
	default:
		if err := s.writer.CreateBank(ctx, bank); err != nil {
			return nil, fmt.Errorf("failed to create bank: %w", err)
		}
	}

	return bank, nil
}

func (s *BankCreatorService) requestToBank(request *CreateBankRequest) *models.Bank {
	// BankID is now required, use it directly
	bankID := request.BankID

	var bankGroupID *uuid.UUID
	if request.BankGroupID != nil && *request.BankGroupID != "" {
		parsed, err := uuid.Parse(*request.BankGroupID)
		if err != nil {
			slog.Warn("invalid bank group ID format, ignoring bank group assignment",
				"error", err.Error(),
				"bank_group_id", *request.BankGroupID,
				"bank_name", request.Name,
				"bank_will_be_created_without_group", true,
			)
			// Continue without bankGroupID (backwards compatibility)
		} else {
			bankGroupID = &parsed
		}
	}

	return &models.Bank{
		BankID:                 bankID,
		Name:                   request.Name,
		BankCodes:              request.BankCodes,
		BIC:                    request.BIC,
		RealName:               request.RealName,
		API:                    request.API,
		APIVersion:             request.APIVersion,
		ASPSP:                  request.ASPSP,
		ProductCode:            request.ProductCode,
		Country:                request.Country,
		BankGroupID:            bankGroupID,
		LogoURL:                request.LogoURL,
		Documentation:          request.Documentation,
		Keywords:               request.Keywords,
		Attribute:              request.Attribute,
		AuthTypeChoiceRequired: request.AuthTypeChoiceRequired,
	}
}

func (s *BankCreatorService) buildEnvironmentConfigs(request *CreateBankRequest, bankID string) []*models.BankEnvironmentConfig {
	configs := make([]*models.BankEnvironmentConfig, 0, len(request.Environments))

	for _, env := range request.Environments {
		config := &models.BankEnvironmentConfig{
			BankID:      bankID,
			Environment: models.EnvironmentType(env),
		}

		if request.Configuration != nil {
			s.applyEnvironmentConfig(config, request.Configuration)
		}

		configs = append(configs, config)
	}

	return configs
}

func (s *BankCreatorService) buildConfigurationsConfigs(request *CreateBankRequest, bankID string) []*models.BankEnvironmentConfig {
	configs := make([]*models.BankEnvironmentConfig, 0, len(request.Configurations))

	for env, envConfig := range request.Configurations {
		config := &models.BankEnvironmentConfig{
			BankID:      bankID,
			Environment: models.EnvironmentType(env),
		}

		s.applyEnvironmentConfig(config, envConfig)
		configs = append(configs, config)
	}

	return configs
}

func (s *BankCreatorService) applyEnvironmentConfig(config *models.BankEnvironmentConfig, envConfig *EnvironmentConfig) {
	if envConfig.Enabled != nil {
		config.Enabled = *envConfig.Enabled
	}
	if envConfig.Blocked != nil {
		config.Blocked = *envConfig.Blocked
	}
	config.BlockedText = envConfig.BlockedText
	if envConfig.Risky != nil {
		config.Risky = *envConfig.Risky
	}
	config.RiskyMessage = envConfig.RiskyMessage
	config.SupportsInstantPayments = envConfig.SupportsInstantPayments
	config.InstantPaymentsActivated = envConfig.InstantPaymentsActivated
	config.InstantPaymentsLimit = envConfig.InstantPaymentsLimit
	config.OkStatusCodesSimplePayment = envConfig.OkStatusCodesSimplePayment
	config.OkStatusCodesInstantPayment = envConfig.OkStatusCodesInstantPayment
	config.OkStatusCodesPeriodicPayment = envConfig.OkStatusCodesPeriodicPayment
	config.EnabledPeriodicPayment = envConfig.EnabledPeriodicPayment
	config.FrequencyPeriodicPayment = envConfig.FrequencyPeriodicPayment
	config.ConfigPeriodicPayment = envConfig.ConfigPeriodicPayment
	if envConfig.AppAuthSetupRequired != nil {
		config.AppAuthSetupRequired = *envConfig.AppAuthSetupRequired
	}
}
