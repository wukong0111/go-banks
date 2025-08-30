package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wukong0111/go-banks/internal/models"
)

type PostgresBankRepository struct {
	db *pgxpool.Pool
}

func NewPostgresBankRepository(db *pgxpool.Pool) *PostgresBankRepository {
	return &PostgresBankRepository{db: db}
}

func (r *PostgresBankRepository) GetBanks(ctx context.Context, filters *BankFilters) ([]models.Bank, *models.Pagination, error) {
	// Build WHERE clause and arguments
	var whereConditions []string
	var args []any
	argIndex := 1

	// Environment filter
	if filters.Environment != "" && filters.Environment != "all" {
		whereConditions = append(whereConditions, fmt.Sprintf("EXISTS (SELECT 1 FROM bank_environment_configs bec WHERE bec.bank_id = b.bank_id AND bec.environment = $%d)", argIndex))
		args = append(args, filters.Environment)
		argIndex++
	}

	// Name filter (partial match)
	if filters.Name != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("b.name ILIKE $%d", argIndex))
		args = append(args, "%"+filters.Name+"%")
		argIndex++
	}

	// API filter
	if filters.API != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("b.api = $%d", argIndex))
		args = append(args, filters.API)
		argIndex++
	}

	// Country filter
	if filters.Country != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("b.country = $%d", argIndex))
		args = append(args, filters.Country)
		argIndex++
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM banks b " + whereClause
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, nil, fmt.Errorf("failed to count banks: %w", err)
	}

	// Calculate pagination
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}

	totalPages := (total + filters.Limit - 1) / filters.Limit
	offset := (filters.Page - 1) * filters.Limit

	pagination := &models.Pagination{
		Page:       filters.Page,
		Limit:      filters.Limit,
		Total:      total,
		TotalPages: totalPages,
	}

	// Get banks with pagination
	query := fmt.Sprintf(`
		SELECT 
			bank_id, name, bank_codes, bic, real_name, api, api_version, 
			aspsp, product_code, country, bank_group_id, logo_url, 
			documentation, keywords, attribute, auth_type_choice_required,
			created_at, updated_at
		FROM banks b
		%s
		ORDER BY b.name
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query banks: %w", err)
	}
	defer rows.Close()

	var banks []models.Bank
	for rows.Next() {
		var bank models.Bank
		err := rows.Scan(
			&bank.BankID, &bank.Name, &bank.BankCodes, &bank.BIC, &bank.RealName,
			&bank.API, &bank.APIVersion, &bank.ASPSP, &bank.ProductCode, &bank.Country,
			&bank.BankGroupID, &bank.LogoURL, &bank.Documentation, &bank.Keywords,
			&bank.Attribute, &bank.AuthTypeChoiceRequired, &bank.CreatedAt, &bank.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan bank: %w", err)
		}
		banks = append(banks, bank)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating bank rows: %w", err)
	}

	return banks, pagination, nil
}

func (r *PostgresBankRepository) GetBankByID(ctx context.Context, bankID string) (*models.Bank, error) {
	query := `
		SELECT 
			bank_id, name, bank_codes, bic, real_name, api, api_version, 
			aspsp, product_code, country, bank_group_id, logo_url, 
			documentation, keywords, attribute, auth_type_choice_required,
			created_at, updated_at
		FROM banks 
		WHERE bank_id = $1
	`

	var bank models.Bank
	err := r.db.QueryRow(ctx, query, bankID).Scan(
		&bank.BankID, &bank.Name, &bank.BankCodes, &bank.BIC, &bank.RealName,
		&bank.API, &bank.APIVersion, &bank.ASPSP, &bank.ProductCode, &bank.Country,
		&bank.BankGroupID, &bank.LogoURL, &bank.Documentation, &bank.Keywords,
		&bank.Attribute, &bank.AuthTypeChoiceRequired, &bank.CreatedAt, &bank.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get bank by ID: %w", err)
	}

	return &bank, nil
}

func (r *PostgresBankRepository) GetBankEnvironmentConfigs(ctx context.Context, bankID, environment string) (map[string]*models.BankEnvironmentConfig, error) {
	query := `
		SELECT 
			bank_id, environment, enabled, blocked, blocked_text, risky, risky_message,
			supports_instant_payments, instant_payments_activated, instant_payments_limit,
			ok_status_codes_simple_payment, ok_status_codes_instant_payment, 
			ok_status_codes_periodic_payment, enabled_periodic_payment, 
			frequency_periodic_payment, config_periodic_payment, app_auth_setup_required,
			created_at, updated_at
		FROM bank_environment_configs 
		WHERE bank_id = $1
	`

	args := []any{bankID}

	// If specific environment is requested, add filter
	if environment != "" {
		query += " AND environment = $2"
		args = append(args, environment)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query bank environment configs: %w", err)
	}
	defer rows.Close()

	configs := make(map[string]*models.BankEnvironmentConfig)
	for rows.Next() {
		var config models.BankEnvironmentConfig
		err := rows.Scan(
			&config.BankID, &config.Environment, &config.Enabled, &config.Blocked,
			&config.BlockedText, &config.Risky, &config.RiskyMessage,
			&config.SupportsInstantPayments, &config.InstantPaymentsActivated,
			&config.InstantPaymentsLimit, &config.OkStatusCodesSimplePayment,
			&config.OkStatusCodesInstantPayment, &config.OkStatusCodesPeriodicPayment,
			&config.EnabledPeriodicPayment, &config.FrequencyPeriodicPayment,
			&config.ConfigPeriodicPayment, &config.AppAuthSetupRequired,
			&config.CreatedAt, &config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bank environment config: %w", err)
		}
		configs[string(config.Environment)] = &config
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment config rows: %w", err)
	}

	return configs, nil
}

func (r *PostgresBankRepository) GetAvailableFilters(ctx context.Context) (*models.BankFilters, error) {
	filters := &models.BankFilters{}

	// Get countries with counts
	countriesQuery := `
		SELECT country, COUNT(*) as count
		FROM banks
		GROUP BY country
		ORDER BY country
	`

	countryRows, err := r.db.Query(ctx, countriesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query countries: %w", err)
	}
	defer countryRows.Close()

	for countryRows.Next() {
		var country models.CountryFilter
		if err := countryRows.Scan(&country.Code, &country.Count); err != nil {
			return nil, fmt.Errorf("failed to scan country: %w", err)
		}
		// TODO: Add country name mapping if needed
		country.Name = country.Code // Placeholder, could be enhanced with country name lookup
		filters.Countries = append(filters.Countries, country)
	}

	if err := countryRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating countries: %w", err)
	}

	// Get APIs with counts
	apisQuery := `
		SELECT api, COUNT(*) as count
		FROM banks
		GROUP BY api
		ORDER BY api
	`

	apiRows, err := r.db.Query(ctx, apisQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query APIs: %w", err)
	}
	defer apiRows.Close()

	for apiRows.Next() {
		var api models.APIFilter
		if err := apiRows.Scan(&api.Type, &api.Count); err != nil {
			return nil, fmt.Errorf("failed to scan API: %w", err)
		}
		filters.APIs = append(filters.APIs, api)
	}

	if err := apiRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating APIs: %w", err)
	}

	// Get distinct environments
	environmentsQuery := `
		SELECT DISTINCT environment
		FROM bank_environment_configs
		ORDER BY environment
	`

	envRows, err := r.db.Query(ctx, environmentsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query environments: %w", err)
	}
	defer envRows.Close()

	for envRows.Next() {
		var env string
		if err := envRows.Scan(&env); err != nil {
			return nil, fmt.Errorf("failed to scan environment: %w", err)
		}
		filters.Environments = append(filters.Environments, env)
	}

	if err := envRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environments: %w", err)
	}

	// Get bank groups with counts
	bankGroupsQuery := `
		SELECT bg.group_id, bg.name, COUNT(b.bank_id) as count
		FROM bank_groups bg
		LEFT JOIN banks b ON bg.group_id = b.bank_group_id
		GROUP BY bg.group_id, bg.name
		ORDER BY bg.name
	`

	groupRows, err := r.db.Query(ctx, bankGroupsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query bank groups: %w", err)
	}
	defer groupRows.Close()

	for groupRows.Next() {
		var group models.BankGroupFilter
		if err := groupRows.Scan(&group.GroupID, &group.Name, &group.Count); err != nil {
			return nil, fmt.Errorf("failed to scan bank group: %w", err)
		}
		filters.BankGroups = append(filters.BankGroups, group)
	}

	if err := groupRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bank groups: %w", err)
	}

	return filters, nil
}
