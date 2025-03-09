package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ziliscite/video-to-mp3/converter/internal/domain"
)

var (
	ErrDuplicateEntry = errors.New("duplicate")
)

type MetadataWriter interface {
	Insert(ctx context.Context, metadata *domain.Metadata) error
}

type MetadataRepository interface {
	MetadataWriter
}

func NewMetadataRepo(db *pgxpool.Pool) MetadataRepository {
	return &metadataRepo{db: db}
}

type metadataRepo struct {
	db *pgxpool.Pool
}

func (u metadataRepo) Insert(ctx context.Context, metadata *domain.Metadata) error {
	query := `
        INSERT INTO metadata(user_id, file_name, video_key, audio_key) 
        VALUES ($1, $2, $3, $4)
	`

	args := []any{metadata.UserId, metadata.FileName, metadata.VideoKey, metadata.AudioKey}

	if _, err := u.db.Exec(ctx, query, args...); err != nil {
		var pgErr *pgconn.PgError
		switch {
		case errors.As(err, &pgErr) && pgErr.Code == "23505":
			return ErrDuplicateEntry // won't be any duplicate entries since we didn't put unique constraints
		default:
			return fmt.Errorf("something's wrong: %w", err)
		}
	}

	return nil
}
