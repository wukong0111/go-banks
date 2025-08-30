package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/wukong0111/go-banks/internal/models"
	"github.com/wukong0111/go-banks/internal/repository"
)

type UpdateBankRequest struct {
	BankID                 *string                       `json:"bank_id,omitempty"`
	Name                   *string                       `json:"name,omitempty"`
	BankCodes              []string                      `json:"bank_codes,omitempty"`
	API                    *string                       `json:"api,omitempty"`
	APIVersion             *string                       `json:"api_version,omitempty"`
	ASPSP                  *string                       `json:"aspsp,omitempty"`
	Country                *string                       `json:"country,omitempty"`
	AuthTypeChoiceRequired *bool                         `json:"auth_type_choice_required,omitempty"`
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

type UpdateBankResponse struct {
	Bank               *models.Bank                    `json:"bank"`
	EnvironmentConfigs []*models.BankEnvironmentConfig `json:"environment_configs"`
}

type BankUpdater interface {
	UpdateBank(ctx context.Context, bankID string, request *UpdateBankRequest) (*UpdateBankResponse, error)
}

type BankUpdaterService struct {
	writer repository.BankWriter
	reader repository.BankRepository
}

func NewBankUpdaterService(writer repository.BankWriter, reader repository.BankRepository) *BankUpdaterService {
	return &BankUpdaterService{
		writer: writer,
		reader: reader,
	}
}

func (s *BankUpdaterService) UpdateBank(ctx context.Context, bankID string, request *UpdateBankRequest) (*UpdateBankResponse, error) {
	// Verify bank exists first
	existingBank, err := s.reader.GetBankByID(ctx, bankID)
	if err != nil {
		return nil, fmt.Errorf("bank not found: %w", err)
	}

	// Build updated bank from request
	updatedBank, err := s.requestToBank(existingBank, request)
	if err != nil {
		return nil, err
	}

	// Handle different update scenarios
	switch {
	case request.Environments != nil || request.Configuration != nil:
		configs := s.buildEnvironmentConfigs(request, bankID)
		if err := s.writer.UpdateBankWithEnvironments(ctx, updatedBank, configs); err != nil {
			return nil, fmt.Errorf("failed to update bank with environments: %w", err)
		}
		return &UpdateBankResponse{
			Bank:               updatedBank,
			EnvironmentConfigs: configs,
		}, nil

	case request.Configurations != nil:
		configs := s.buildConfigurationsConfigs(request, bankID)
		if err := s.writer.UpdateBankWithEnvironments(ctx, updatedBank, configs); err != nil {
			return nil, fmt.Errorf("failed to update bank with configurations: %w", err)
		}
		return &UpdateBankResponse{
			Bank:               updatedBank,
			EnvironmentConfigs: configs,
		}, nil

	default:
		if err := s.writer.UpdateBank(ctx, updatedBank); err != nil {
			return nil, fmt.Errorf("failed to update bank: %w", err)
		}
		return &UpdateBankResponse{
			Bank:               updatedBank,
			EnvironmentConfigs: nil,
		}, nil
	}
}

func (s *BankUpdaterService) requestToBank(existing *models.Bank, request *UpdateBankRequest) (*models.Bank, error) {
	// Start with existing bank data
	updated := &models.Bank{
		BankID:                 existing.BankID,
		Name:                   existing.Name,
		BankCodes:              existing.BankCodes,
		BIC:                    existing.BIC,
		RealName:               existing.RealName,
		API:                    existing.API,
		APIVersion:             existing.APIVersion,
		ASPSP:                  existing.ASPSP,
		ProductCode:            existing.ProductCode,
		Country:                existing.Country,
		BankGroupID:            existing.BankGroupID,
		LogoURL:                existing.LogoURL,
		Documentation:          existing.Documentation,
		Keywords:               existing.Keywords,
		Attribute:              existing.Attribute,
		AuthTypeChoiceRequired: existing.AuthTypeChoiceRequired,
	}

	// Apply updates only for provided fields
	if request.BankID != nil {
		updated.BankID = *request.BankID
	}
	if request.Name != nil {
		updated.Name = *request.Name
	}
	if request.BankCodes != nil {
		updated.BankCodes = request.BankCodes
	}
	if request.API != nil {
		updated.API = *request.API
	}
	if request.APIVersion != nil {
		updated.APIVersion = *request.APIVersion
	}
	if request.ASPSP != nil {
		updated.ASPSP = *request.ASPSP
	}
	if request.Country != nil {
		updated.Country = *request.Country
	}
	if request.AuthTypeChoiceRequired != nil {
		updated.AuthTypeChoiceRequired = *request.AuthTypeChoiceRequired
	}
	if request.BIC != nil {
		updated.BIC = request.BIC
	}
	if request.RealName != nil {
		updated.RealName = request.RealName
	}
	if request.ProductCode != nil {
		updated.ProductCode = request.ProductCode
	}
	if request.LogoURL != nil {
		updated.LogoURL = request.LogoURL
	}
	if request.Documentation != nil {
		updated.Documentation = request.Documentation
	}
	if request.Keywords != nil {
		updated.Keywords = request.Keywords
	}
	if request.Attribute != nil {
		updated.Attribute = request.Attribute
	}

	// Handle BankGroupID update
	if request.BankGroupID != nil {
		bankGroupID, err := s.parseBankGroupID(request.BankGroupID)
		if err != nil {
			return nil, err
		}
		updated.BankGroupID = bankGroupID
	}

	return updated, nil
}

// parseBankGroupID safely parses an optional BankGroupID string pointer into a UUID pointer
func (s *BankUpdaterService) parseBankGroupID(bankGroupID *string) (*uuid.UUID, error) {
	// Return nil if not provided
	if bankGroupID == nil {
		return nil, nil
	}

	// Trim whitespace and check if empty
	trimmedID := strings.TrimSpace(*bankGroupID)
	if trimmedID == "" {
		return nil, nil
	}

	// Parse as UUID
	parsed, err := uuid.Parse(trimmedID)
	if err != nil {
		return nil, fmt.Errorf("invalid bank_group_id format: %w", err)
	}

	return &parsed, nil
}

func (s *BankUpdaterService) buildEnvironmentConfigs(request *UpdateBankRequest, bankID string) []*models.BankEnvironmentConfig {
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

func (s *BankUpdaterService) buildConfigurationsConfigs(request *UpdateBankRequest, bankID string) []*models.BankEnvironmentConfig {
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

func (s *BankUpdaterService) applyEnvironmentConfig(config *models.BankEnvironmentConfig, envConfig *EnvironmentConfig) {
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
