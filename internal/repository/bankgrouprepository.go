package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wukong0111/go-banks/internal/models"
)

// PostgresBankGroupRepository implements BankGroupRepository interface
type PostgresBankGroupRepository struct {
	db *pgxpool.Pool
}

// NewPostgresBankGroupRepository creates a new PostgresBankGroupRepository instance
func NewPostgresBankGroupRepository(db *pgxpool.Pool) *PostgresBankGroupRepository {
	return &PostgresBankGroupRepository{db: db}
}

// GetBankGroups retrieves all bank groups from the database
func (r *PostgresBankGroupRepository) GetBankGroups(ctx context.Context) ([]models.BankGroup, error) {
	query := `
		SELECT group_id, name, description, logo_url, website, created_at, updated_at
		FROM bank_groups 
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bankGroups []models.BankGroup
	for rows.Next() {
		var bg models.BankGroup
		err := rows.Scan(
			&bg.GroupID,
			&bg.Name,
			&bg.Description,
			&bg.LogoURL,
			&bg.Website,
			&bg.CreatedAt,
			&bg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bankGroups = append(bankGroups, bg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return bankGroups, nil
}
