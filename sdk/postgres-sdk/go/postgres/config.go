package postgres

import "fmt"

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func NewConfig(host string, port int, user, password, database string) *Config {
	return &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Database)
}
