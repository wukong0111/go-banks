package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wukong0111/go-banks/internal/models"
)

type PostgresBankWriter struct {
	db *pgxpool.Pool
}

func NewPostgresBankWriter(db *pgxpool.Pool) *PostgresBankWriter {
	return &PostgresBankWriter{db: db}
}

func (w *PostgresBankWriter) CreateBank(ctx context.Context, bank *models.Bank) error {
	if bank.BankID == "" {
		bank.BankID = uuid.New().String()
	}

	query := `
		INSERT INTO banks (
			bank_id, name, bank_codes, bic, real_name, api, api_version,
			aspsp, product_code, country, bank_group_id, logo_url,
			documentation, keywords, attribute, auth_type_choice_required
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err := w.db.Exec(ctx, query,
		bank.BankID, bank.Name, bank.BankCodes, bank.BIC, bank.RealName,
		bank.API, bank.APIVersion, bank.ASPSP, bank.ProductCode, bank.Country,
		bank.BankGroupID, bank.LogoURL, bank.Documentation, bank.Keywords,
		bank.Attribute, bank.AuthTypeChoiceRequired,
	)

	if err != nil {
		return fmt.Errorf("failed to create bank: %w", err)
	}

	return nil
}

func (w *PostgresBankWriter) CreateBankWithEnvironments(ctx context.Context, bank *models.Bank, configs []*models.BankEnvironmentConfig) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if bank.BankID == "" {
		bank.BankID = uuid.New().String()
	}

	bankQuery := `
		INSERT INTO banks (
			bank_id, name, bank_codes, bic, real_name, api, api_version,
			aspsp, product_code, country, bank_group_id, logo_url,
			documentation, keywords, attribute, auth_type_choice_required
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)
	`

	_, err = tx.Exec(ctx, bankQuery,
		bank.BankID, bank.Name, bank.BankCodes, bank.BIC, bank.RealName,
		bank.API, bank.APIVersion, bank.ASPSP, bank.ProductCode, bank.Country,
		bank.BankGroupID, bank.LogoURL, bank.Documentation, bank.Keywords,
		bank.Attribute, bank.AuthTypeChoiceRequired,
	)

	if err != nil {
		return fmt.Errorf("failed to create bank: %w", err)
	}

	configQuery := `
		INSERT INTO bank_environment_configs (
			bank_id, environment, enabled, blocked, blocked_text, risky, risky_message,
			supports_instant_payments, instant_payments_activated, instant_payments_limit,
			ok_status_codes_simple_payment, ok_status_codes_instant_payment,
			ok_status_codes_periodic_payment, enabled_periodic_payment,
			frequency_periodic_payment, config_periodic_payment, app_auth_setup_required
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	for _, config := range configs {
		config.BankID = bank.BankID
		_, err = tx.Exec(ctx, configQuery,
			config.BankID, config.Environment, config.Enabled, config.Blocked,
			config.BlockedText, config.Risky, config.RiskyMessage,
			config.SupportsInstantPayments, config.InstantPaymentsActivated,
			config.InstantPaymentsLimit, config.OkStatusCodesSimplePayment,
			config.OkStatusCodesInstantPayment, config.OkStatusCodesPeriodicPayment,
			config.EnabledPeriodicPayment, config.FrequencyPeriodicPayment,
			config.ConfigPeriodicPayment, config.AppAuthSetupRequired,
		)

		if err != nil {
			return fmt.Errorf("failed to create environment config for %s: %w", config.Environment, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (w *PostgresBankWriter) UpdateBank(ctx context.Context, bank *models.Bank) error {
	query := `
		UPDATE banks SET
			name = $2, bank_codes = $3, bic = $4, real_name = $5, api = $6,
			api_version = $7, aspsp = $8, product_code = $9, country = $10,
			bank_group_id = $11, logo_url = $12, documentation = $13,
			keywords = $14, attribute = $15, auth_type_choice_required = $16,
			updated_at = CURRENT_TIMESTAMP
		WHERE bank_id = $1
	`

	result, err := w.db.Exec(ctx, query,
		bank.BankID, bank.Name, bank.BankCodes, bank.BIC, bank.RealName,
		bank.API, bank.APIVersion, bank.ASPSP, bank.ProductCode, bank.Country,
		bank.BankGroupID, bank.LogoURL, bank.Documentation, bank.Keywords,
		bank.Attribute, bank.AuthTypeChoiceRequired,
	)

	if err != nil {
		return fmt.Errorf("failed to update bank: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("bank with ID '%s' not found", bank.BankID)
	}

	return nil
}

func (w *PostgresBankWriter) UpdateBankWithEnvironments(ctx context.Context, bank *models.Bank, configs []*models.BankEnvironmentConfig) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Update bank table
	bankQuery := `
		UPDATE banks SET
			name = $2, bank_codes = $3, bic = $4, real_name = $5, api = $6,
			api_version = $7, aspsp = $8, product_code = $9, country = $10,
			bank_group_id = $11, logo_url = $12, documentation = $13,
			keywords = $14, attribute = $15, auth_type_choice_required = $16,
			updated_at = CURRENT_TIMESTAMP
		WHERE bank_id = $1
	`

	result, err := tx.Exec(ctx, bankQuery,
		bank.BankID, bank.Name, bank.BankCodes, bank.BIC, bank.RealName,
		bank.API, bank.APIVersion, bank.ASPSP, bank.ProductCode, bank.Country,
		bank.BankGroupID, bank.LogoURL, bank.Documentation, bank.Keywords,
		bank.Attribute, bank.AuthTypeChoiceRequired,
	)

	if err != nil {
		return fmt.Errorf("failed to update bank: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("bank with ID '%s' not found", bank.BankID)
	}

	// Delete existing environment configs for this bank
	deleteQuery := "DELETE FROM bank_environment_configs WHERE bank_id = $1"
	if _, err = tx.Exec(ctx, deleteQuery, bank.BankID); err != nil {
		return fmt.Errorf("failed to delete existing environment configs: %w", err)
	}

	// Insert new environment configs
	configQuery := `
		INSERT INTO bank_environment_configs (
			bank_id, environment, enabled, blocked, blocked_text, risky, risky_message,
			supports_instant_payments, instant_payments_activated, instant_payments_limit,
			ok_status_codes_simple_payment, ok_status_codes_instant_payment,
			ok_status_codes_periodic_payment, enabled_periodic_payment,
			frequency_periodic_payment, config_periodic_payment, app_auth_setup_required
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	for _, config := range configs {
		config.BankID = bank.BankID
		_, err = tx.Exec(ctx, configQuery,
			config.BankID, config.Environment, config.Enabled, config.Blocked,
			config.BlockedText, config.Risky, config.RiskyMessage,
			config.SupportsInstantPayments, config.InstantPaymentsActivated,
			config.InstantPaymentsLimit, config.OkStatusCodesSimplePayment,
			config.OkStatusCodesInstantPayment, config.OkStatusCodesPeriodicPayment,
			config.EnabledPeriodicPayment, config.FrequencyPeriodicPayment,
			config.ConfigPeriodicPayment, config.AppAuthSetupRequired,
		)

		if err != nil {
			return fmt.Errorf("failed to create environment config for %s: %w", config.Environment, err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
