package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wukong0111/go-banks/internal/models"
)

type BankRepository struct {
	db *pgxpool.Pool
}

type BankFilters struct {
	Environment string
	Name        string
	API         string
	Country     string
	Page        int
	Limit       int
}

func NewBankRepository(db *pgxpool.Pool) *BankRepository {
	return &BankRepository{db: db}
}

func (r *BankRepository) GetBanks(ctx context.Context, filters BankFilters) ([]models.Bank, *models.Pagination, error) {
	// Build WHERE clause and arguments
	var whereConditions []string
	var args []interface{}
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
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM banks b %s", whereClause)
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
