package migrations

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"eqray-go-starter/app/facades"
	"eqray-go-starter/app/modules/base/support"
)

type M20261017000001CreateSystemUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M20261017000001CreateSystemUsersTable) Signature() string {
	return "20261017000001_create_system_users_table"
}

// Up 建 system_users 表并写入初始管理员（docs 6.3）。
// 唯一键直接建在业务键上：账号不复用，软删除后同名账号不能再创建。
func (r *M20261017000001CreateSystemUsersTable) Up() error {
	if facades.Schema().HasTable("system_users") {
		return nil
	}

	if err := facades.DB().Statement(`
CREATE TABLE system_users (
  id            bigint       NOT NULL COMMENT '主键',
  username      varchar(64)  NOT NULL COMMENT '登录名，唯一',
  password      varchar(255) NOT NULL COMMENT 'bcrypt 哈希',
  nickname      varchar(64)  NOT NULL DEFAULT '' COMMENT '昵称',
  remark        varchar(500) NOT NULL DEFAULT '' COMMENT '备注',
  email         varchar(128) NULL COMMENT '邮箱',
  mobile        varchar(20)  NULL COMMENT '手机号',
  sex           tinyint      NOT NULL DEFAULT 0 COMMENT '性别：0 未知 / 1 男 / 2 女',
  avatar        varchar(512) NOT NULL DEFAULT '' COMMENT '头像地址',
  status        tinyint      NOT NULL DEFAULT 1 COMMENT '状态：0 禁用 / 1 启用',
  login_ip      varchar(50)  NOT NULL DEFAULT '' COMMENT '最后登录 IP',
  last_login_at datetime     NULL COMMENT '最后登录时间',
  creator       varchar(64)  NOT NULL DEFAULT '' COMMENT '创建人',
  created_at    datetime     NULL COMMENT '创建时间',
  updater       varchar(64)  NOT NULL DEFAULT '' COMMENT '更新人',
  updated_at    datetime     NULL COMMENT '更新时间',
  deleted_at    datetime     NULL COMMENT '删除时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username),
  UNIQUE KEY uk_email (email),
  UNIQUE KEY uk_mobile (mobile)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台用户表'`); err != nil {
		return err
	}

	return r.seedAdmin()
}

// seedAdmin 写入初始管理员 admin / admin123，上线前必须修改密码。
func (r *M20261017000001CreateSystemUsersTable) seedAdmin() error {
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now()

	_, err = facades.DB().Table("system_users").Insert(map[string]any{
		"id":         support.NewID(),
		"username":   "admin",
		"password":   string(hash),
		"nickname":   "超级管理员",
		"status":     1,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	})

	return err
}

// Down Reverse the migrations.
func (r *M20261017000001CreateSystemUsersTable) Down() error {
	return facades.Schema().DropIfExists("system_users")
}
