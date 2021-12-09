package distlock

import (
	"database/sql"
	"fmt"
	"time"
)

const mysqlCreateTableSQL = `
CREATE TABLE IF NOT EXISTS %s (
	id        VARCHAR(255) PRIMARY KEY,
	owner     VARCHAR(255) NOT NULL DEFAULT '',
	expire_at BIGINT NOT NULL DEFAULT 0
);
`

const mysqlLockSQL = `
INSERT INTO %s (id, owner, expire_at) VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
owner = IF(expire_at < ?, VALUES(owner), owner),
expire_at = IF(expire_at < ?, VALUES(expire_at), expire_at);
`

const mysqlUnlockSQL = `DELETE FROM %s WHERE id = ? AND owner = ? AND expire_at >= ?;`

type mysqlProvider struct {
	db         *sql.DB
	table      string
	lockStmt   *sql.Stmt
	unlockStmt *sql.Stmt
}

func NewMySQLProvider(db *sql.DB, tableName string) (Provider, error) {
	provider := &mysqlProvider{
		db:    db,
		table: tableName,
	}
	if err := provider.CreateTable(); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	lockStmt, err := db.Prepare(provider.formatSQL(mysqlLockSQL))
	if err != nil {
		return nil, fmt.Errorf("prepare lock statement: %w", err)
	}
	provider.lockStmt = lockStmt

	unlockStmt, err := db.Prepare(provider.formatSQL(mysqlUnlockSQL))
	if err != nil {
		return nil, fmt.Errorf("prepare unlock statement: %w", err)
	}
	provider.unlockStmt = unlockStmt

	return provider, nil
}

func (p *mysqlProvider) formatSQL(sqlTemplate string) string {
	return fmt.Sprintf(sqlTemplate, p.table)
}

func (p *mysqlProvider) CreateTable() error {
	_, err := p.db.Exec(p.formatSQL(mysqlCreateTableSQL))
	return err
}

func (p *mysqlProvider) Name() string {
	return "mysql"
}

func (p *mysqlProvider) Lock(lock LockInfo) error {
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

func (p *mysqlProvider) Unlock(lock LockInfo) error {
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
