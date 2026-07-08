package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"smart-cabinet/internal/config"
	"smart-cabinet/internal/model"
)

// DB 数据库操作封装
type DB struct {
	conn *sql.DB
}

// New 创建数据库连接并初始化表结构
func New(cfg config.DatabaseConfig) (*DB, error) {
	conn, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("数据库 ping 失败: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.initTables(); err != nil {
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	return db, nil
}

// initTables 自动建表
func (d *DB) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS cabinet_records (
			id         BIGINT AUTO_INCREMENT PRIMARY KEY,
			code       VARCHAR(4)  NOT NULL COMMENT '4位存取密码',
			action     ENUM('store','retrieve') NOT NULL COMMENT '操作类型',
			created_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='存取记录表'`,

		`CREATE TABLE IF NOT EXISTS current_item (
			id        BIGINT AUTO_INCREMENT PRIMARY KEY,
			code      VARCHAR(4)  NOT NULL COMMENT '当前外卖的取件密码',
			stored_at DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '存入时间',
			status    ENUM('stored','retrieved') NOT NULL DEFAULT 'stored' COMMENT '状态'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='当前柜内外卖'`,
	}

	for _, q := range queries {
		if _, err := d.conn.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// InsertRecord 插入存取记录
func (d *DB) InsertRecord(code, action string) error {
	_, err := d.conn.Exec(
		"INSERT INTO cabinet_records (code, action) VALUES (?, ?)",
		code, action,
	)
	return err
}

// SetCurrentItem 设置当前柜内外卖（存入时调用）
func (d *DB) SetCurrentItem(code string) error {
	// 先清空旧记录
	_, _ = d.conn.Exec("UPDATE current_item SET status='retrieved' WHERE status='stored'")
	_, err := d.conn.Exec(
		"INSERT INTO current_item (code, status) VALUES (?, 'stored')",
		code,
	)
	return err
}

// GetCurrentItem 获取当前柜内外卖
func (d *DB) GetCurrentItem() (*model.CurrentItem, error) {
	item := &model.CurrentItem{}
	err := d.conn.QueryRow(
		"SELECT id, code, stored_at, status FROM current_item WHERE status='stored' LIMIT 1",
	).Scan(&item.ID, &item.Code, &item.StoredAt, &item.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

// MarkRetrieved 标记外卖已取走
func (d *DB) MarkRetrieved(id int64) error {
	_, err := d.conn.Exec(
		"UPDATE current_item SET status='retrieved' WHERE id=?",
		id,
	)
	return err
}

// VerifyCode 验证取件码是否匹配
func (d *DB) VerifyCode(code string) (bool, error) {
	var count int
	err := d.conn.QueryRow(
		"SELECT COUNT(*) FROM current_item WHERE code=? AND status='stored'",
		code,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	return d.conn.Close()
}
