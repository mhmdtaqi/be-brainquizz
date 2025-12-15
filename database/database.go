package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Joko206/UAS_PWEB1/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Global database instance
var DB *gorm.DB

// Database configuration struct
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// GetDatabaseConfig reads database configuration from environment variables
func GetDatabaseConfig() *Config {
	config := &Config{
		Host:     getEnv("DB_HOST", ""),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", ""),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", ""),
		SSLMode:  getEnv("DB_SSLMODE", "require"),
		TimeZone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
	}
	return config
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// BuildDSN creates database connection string from config
func (c *Config) BuildDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode, c.TimeZone)
}

// Dsn contains the Data Source Name for PostgreSQL connection (for backward compatibility)
var Dsn string

// InitDB initializes database connection with optimized settings
func InitDB() (*gorm.DB, error) {
	// Get database configuration
	config := GetDatabaseConfig()
	dsn := config.BuildDSN()

	// Set global Dsn for backward compatibility
	Dsn = dsn

	// Configure GORM with optimized settings
	gormConfig := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent), // Reduce logging overhead in production
		PrepareStmt:                              true,                                  // Enable prepared statement caching
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for optimal performance
	sqlDB.SetMaxOpenConns(getEnvAsInt("DB_MAX_OPEN_CONNS", 25))                                     // Maximum number of open connections
	sqlDB.SetMaxIdleConns(getEnvAsInt("DB_MAX_IDLE_CONNS", 10))                                     // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 300)) * time.Second) // 5 minutes
	sqlDB.SetConnMaxIdleTime(time.Duration(getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 60)) * time.Second) // 1 minute

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run AutoMigrate to ensure the database schema is up to date
	if err := db.AutoMigrate(
		&models.Users{},
		&models.Kategori_Soal{},
		&models.Tingkatan{},
		&models.Kelas{},
		&models.Kuis{},
		&models.Soal{},
		&models.Pendidikan{},
		&models.Hasil_Kuis{},
		&models.SoalAnswer{},
		&models.Kelas_Pengguna{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Printf("Database connected successfully with %d max open connections and %d max idle connections",
		getEnvAsInt("DB_MAX_OPEN_CONNS", 25), getEnvAsInt("DB_MAX_IDLE_CONNS", 10))

	return db, nil
}

// GetDBConnection returns the global database instance (singleton pattern)
func GetDBConnection() (*gorm.DB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized. Call InitDB() first")
	}
	return DB, nil
}

// InitializeDatabase initializes the global database instance
func InitializeDatabase() error {
	db, err := InitDB()
	if err != nil {
		return err
	}
	DB = db
	return nil
}

// CloseDB closes the database connection gracefully
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
