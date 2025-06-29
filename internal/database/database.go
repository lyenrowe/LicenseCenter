package database

import (
	"fmt"
	"time"

	"github.com/lyenrowe/LicenseCenter/internal/config"
	"github.com/lyenrowe/LicenseCenter/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

var DB *Database

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) error {
	var db *gorm.DB
	var err error

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// 根据驱动类型连接数据库
	switch cfg.Driver {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), gormConfig)
	case "mysql":
		db, err = gorm.Open(mysql.Open(cfg.DSN), gormConfig)
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DSN), gormConfig)
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}

	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层sql.DB对象进行连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	DB = &Database{DB: db}
	return nil
}

// AutoMigrate 自动迁移数据库表结构
func (d *Database) AutoMigrate() error {
	return d.DB.AutoMigrate(
		&models.Authorization{},
		&models.License{},
		&models.AdminUser{},
		&models.AdminLog{},
		&models.RSAKey{},
		&models.SystemConfig{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	if DB == nil {
		panic("数据库未初始化")
	}
	return DB.DB
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// IsTableExists 检查表是否存在
func (d *Database) IsTableExists(tableName string) bool {
	return d.DB.Migrator().HasTable(tableName)
}

// CreateIndexes 创建索引
func (d *Database) CreateIndexes() error {
	// 授权码表索引
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_authorizations_code ON authorizations(authorization_code)").Error; err != nil {
		return err
	}
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_authorizations_status ON authorizations(status)").Error; err != nil {
		return err
	}

	// 激活设备表索引
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_licenses_auth_id ON licenses(authorization_id)").Error; err != nil {
		return err
	}
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_licenses_machine_id ON licenses(machine_id)").Error; err != nil {
		return err
	}
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_licenses_status ON licenses(status)").Error; err != nil {
		return err
	}

	// 管理员日志表索引
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_admin_logs_admin_id ON admin_logs(admin_id)").Error; err != nil {
		return err
	}
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_admin_logs_action ON admin_logs(action)").Error; err != nil {
		return err
	}
	if err := d.DB.Exec("CREATE INDEX IF NOT EXISTS idx_admin_logs_created_at ON admin_logs(created_at)").Error; err != nil {
		return err
	}

	return nil
}
