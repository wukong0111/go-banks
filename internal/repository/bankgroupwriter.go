package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wukong0111/go-banks/internal/models"
)

// PostgresBankGroupWriter implements BankGroupWriter interface
type PostgresBankGroupWriter struct {
	db *pgxpool.Pool
}

// NewPostgresBankGroupWriter creates a new PostgresBankGroupWriter instance
func NewPostgresBankGroupWriter(db *pgxpool.Pool) *PostgresBankGroupWriter {
	return &PostgresBankGroupWriter{db: db}
}

// CreateBankGroup inserts a new bank group into the database
func (w *PostgresBankGroupWriter) CreateBankGroup(ctx context.Context, bankGroup *models.BankGroup) error {
	query := `
		INSERT INTO bank_groups (group_id, name, description, logo_url, website, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := w.db.Exec(ctx, query,
		bankGroup.GroupID,
		bankGroup.Name,
		bankGroup.Description,
		bankGroup.LogoURL,
		bankGroup.Website,
		bankGroup.CreatedAt,
		bankGroup.UpdatedAt,
	)

	return err
}

// UpdateBankGroup updates an existing bank group in the database
func (w *PostgresBankGroupWriter) UpdateBankGroup(ctx context.Context, bankGroup *models.BankGroup) error {
	query := `
		UPDATE bank_groups SET
			name = $2, description = $3, logo_url = $4, website = $5,
			updated_at = CURRENT_TIMESTAMP
		WHERE group_id = $1
	`

	result, err := w.db.Exec(ctx, query,
		bankGroup.GroupID,
		bankGroup.Name,
		bankGroup.Description,
		bankGroup.LogoURL,
		bankGroup.Website,
	)

	if err != nil {
		return fmt.Errorf("failed to update bank group: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("bank group with ID '%s' not found", bankGroup.GroupID)
	}

	return nil
}
