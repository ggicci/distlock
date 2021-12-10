package distlock

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	pgCreateTableSQL = `CREATE TABLE IF NOT EXISTS %s (
		id        VARCHAR(255) PRIMARY KEY,
		owner     VARCHAR(255) NOT NULL DEFAULT '',
		expire_at BIGINT NOT NULL DEFAULT 0
	);`

	pgLockSQL = `INSERT INTO %s AS t (id, owner, expire_at) VALUES ($1, $2, $3)
	ON CONFLICT (id) DO UPDATE
	SET owner = $2, expire_at = $3 WHERE t.id = $1 AND t.expire_at < $4;`

	pgUnlockSQL = `DELETE FROM %s WHERE id = $1 AND owner = $2 AND expire_at >= $3;`
)

type postgreSQLProvider mysqlProvider

func NewPostgreSQLProvider(db *sql.DB, tableName string) (Provider, error) {
	provider := &postgreSQLProvider{
		tableName: tableName,
		db:        db,
	}

	return provider, provider.init()
}

func (p *postgreSQLProvider) Name() string {
	return "postgres"
}

func (p *postgreSQLProvider) init() error {
	if _, err := p.db.Exec(formatSQL(pgCreateTableSQL, p.tableName)); err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	db := p.db

	lockStmt, err := db.Prepare(formatSQL(pgLockSQL, p.tableName))
	if err != nil {
		return fmt.Errorf("prepare lock statement: %w", err)
	}
	p.lockStmt = lockStmt

	unlockStmt, err := db.Prepare(formatSQL(pgUnlockSQL, p.tableName))
	if err != nil {
		return fmt.Errorf("prepare unlock statement: %w", err)
	}
	p.unlockStmt = unlockStmt

	return nil
}

func (p *postgreSQLProvider) Lock(lock LockInfo) error {
	now := time.Now()
	expireAt := now.Add(lock.GetLifetime())
	rs, err := p.lockStmt.Exec(
		lock.GetLockId(),
		lock.GetLockOwner(),
		expireAt.UnixNano(),
		now.UnixNano(),
	)
	if err != nil {
		return fmt.Errorf("exec lock statement: %w", err)
	}
	affected, err := rs.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	println("lock affected:", affected)
	if affected == 0 {
		return ErrAlreadyLocked
	}
	return nil
}

func (p *postgreSQLProvider) Unlock(lock LockInfo) error {
	rs, err := p.unlockStmt.Exec(
		lock.GetLockId(),
		lock.GetLockOwner(),
		time.Now().UnixNano(),
	)
	if err != nil {
		return fmt.Errorf("exec unlock statement: %w", err)
	}
	affected, err := rs.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	println("unlock affected:", affected)
	if affected == 0 {
		return ErrNotLocked
	}
	return nil
}
