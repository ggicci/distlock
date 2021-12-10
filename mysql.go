package distlock

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	mysqlCreateTableSQL = `CREATE TABLE IF NOT EXISTS %s (
		id        VARCHAR(255) PRIMARY KEY,
		owner     VARCHAR(255) NOT NULL DEFAULT '',
		expire_at BIGINT NOT NULL DEFAULT 0
	);`

	mysqlLockSQL = `INSERT INTO %s (id, owner, expire_at) VALUES (?, ?, ?)
	ON DUPLICATE KEY UPDATE
	owner = IF(expire_at < ?, VALUES(owner), owner),
	expire_at = IF(expire_at < ?, VALUES(expire_at), expire_at);`

	mysqlUnlockSQL = `DELETE FROM %s WHERE id = ? AND owner = ? AND expire_at >= ?;`
)

type mysqlProvider struct {
	tableName string

	db         *sql.DB
	lockStmt   *sql.Stmt
	unlockStmt *sql.Stmt
}

func NewMySQLProvider(db *sql.DB, tableName string) (Provider, error) {
	provider := &mysqlProvider{
		tableName: tableName,
		db:        db,
	}

	return provider, provider.init()
}

func (p *mysqlProvider) Name() string {
	return "mysql"
}

func (p *mysqlProvider) init() error {
	if _, err := p.db.Exec(formatSQL(mysqlCreateTableSQL, p.tableName)); err != nil {
		return fmt.Errorf("create table: %w", err)
	}
	db := p.db

	lockStmt, err := db.Prepare(formatSQL(mysqlLockSQL, p.tableName))
	if err != nil {
		return fmt.Errorf("prepare lock statement: %w", err)
	}
	p.lockStmt = lockStmt

	unlockStmt, err := db.Prepare(formatSQL(mysqlUnlockSQL, p.tableName))
	if err != nil {
		return fmt.Errorf("prepare unlock statement: %w", err)
	}
	p.unlockStmt = unlockStmt

	return nil
}

func (p *mysqlProvider) Lock(lock NamedLock) error {
	now := time.Now()
	expireAt := now.Add(lock.GetLifetime())
	rs, err := p.lockStmt.Exec(
		lock.GetLockId(),
		lock.GetLockOwner(),
		expireAt.UnixNano(),
		now.UnixNano(),
		now.UnixNano(),
	)
	if err != nil {
		return fmt.Errorf("exec lock statement: %w", err)
	}
	affected, err := rs.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}
	if affected == 0 {
		return ErrAlreadyLocked
	}
	return nil
}

func (p *mysqlProvider) Unlock(lock NamedLock) error {
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
	if affected == 0 {
		return ErrNotLocked
	}
	return nil
}

func formatSQL(sqlTemplate, tableName string) string {
	return fmt.Sprintf(sqlTemplate, tableName)
}
