package repository

import (
	"context"

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
