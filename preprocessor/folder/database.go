package preprocessor_folder

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/starlinglab/integrity-v2/database"
)

func initFileStatusTableIfNotExists(connPool *pgxpool.Pool) error {
	var exists bool
	err := connPool.QueryRow(
		db.GetDatabaseContext(),
		"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'file_status');",
	).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		_, err = connPool.Exec(
			db.GetDatabaseContext(),
			FILE_STATUS_TABLE,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func initDbTableIfNotExists(connPool *pgxpool.Pool) error {
	err := initFileStatusTableIfNotExists(connPool)
	return err
}

type FileQueryResult struct {
	Status       *string
	Cid          *string
	ErrorMessage *string
}

func queryIfFileExists(connPool *pgxpool.Pool, filePath string) (*FileQueryResult, error) {
	var result FileQueryResult
	err := connPool.QueryRow(
		db.GetDatabaseContext(),
		"SELECT status, cid, error FROM file_status WHERE file_path = $1;",
		filePath,
	).Scan(&result.Status, &result.Cid, &result.ErrorMessage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func setFileStatusFound(connPool *pgxpool.Pool, filePath string) error {
	_, err := connPool.Exec(
		db.GetDatabaseContext(),
		"INSERT INTO file_status (file_path, status, created_at, updated_at) VALUES ($1, $2, $3, $4);",
		filePath,
		FileStatusFound,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	return err
}

func setFileStatusUploading(connPool *pgxpool.Pool, filePath string, sha256 string) error {
	_, err := connPool.Exec(
		db.GetDatabaseContext(),
		"UPDATE file_status SET status = $1, sha256 = $2, updated_at = $3 WHERE file_path = $4;",
		FileStatusUploading,
		sha256,
		time.Now().UTC(),
		filePath,
	)
	return err
}

func setFileStatusDone(connPool *pgxpool.Pool, filePath string, cid string) error {
	_, err := connPool.Exec(
		db.GetDatabaseContext(),
		"UPDATE file_status SET status = $1, cid = $2, updated_at = $3 WHERE file_path = $4;",
		FileStatusSuccess,
		cid,
		time.Now().UTC(),
		filePath,
	)
	return err
}

func setFileStatusError(connPool *pgxpool.Pool, filePath string, errorMessage string) error {
	_, err := connPool.Exec(
		db.GetDatabaseContext(),
		"UPDATE file_status SET status = $1, error = $2, updated_at = $3 WHERE file_path = $4;",
		FileStatusError,
		errorMessage,
		time.Now().UTC(),
		filePath,
	)
	return err
}
