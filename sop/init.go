package sop

import (
	"database/sql"
	_ "embed"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/xmzail/ai-carrier-dev/carriercore/common/db"
	csql "github.com/xmzail/ai-carrier-dev/carriercore/common/sql"
	api "github.com/xmzail/ai-carrier-dev/sdk/mountcore-sdk/go"
)

type Config struct {
	Port string
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	config Config
	pgDB   *sql.DB
)

func initConfig() {
	config = Config{
		Port: getEnv("PORT", "2428"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func initDB() error {
	dsn := db.GetDSN()

	log.Println("Initializing Postgres connection for SOP module...")

	var err error
	pgDB, err = sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	if err := pgDB.Ping(); err != nil {
		return err
	}

	pgDB.SetMaxOpenConns(25)
	pgDB.SetMaxIdleConns(5)

	log.Println("Postgres connected successfully for SOP module")

	return nil
}

func Close() {
	if pgDB != nil {
		pgDB.Close()
	}
}

func Init() error {
	initConfig()
	if err := initDB(); err != nil {
		return err
	}
	return nil
}

var pgDBExported *sql.DB

func GetDB() *sql.DB {
	return pgDB
}

func LoadSQLConfig() error {
	return csql.LoadSQLConfigData([]byte{}, "mounts/sop/sql_sop.json")
}

//go:embed api_sop.json
var apiSOPData []byte

func getHandlerMap() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GetSOPListHandler": GetSOPListHandler,
		"GetSOPHandler":     GetSOPHandler,
		"CreateSOPHandler":  CreateSOPHandler,
		"UpdateSOPHandler":  UpdateSOPHandler,
		"DeleteSOPHandler":  DeleteSOPHandler,
	}
}

func RegisterRoutes(router gin.IRouter) {
	if err := api.RegisterRoutesFromJSONData(router, apiSOPData, "sop/api_sop.json", getHandlerMap()); err != nil {
		log.Fatalf("Failed to register SOP routes: %v", err)
	}
}
