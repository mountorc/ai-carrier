package role

import (
	"database/sql"
	_ "embed"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/trae/autoFlow/carriercore/common/db"
	csql "github.com/trae/autoFlow/carriercore/common/sql"
	api "github.com/trae/autoFlow/sdk/mountcore-sdk/go"
)

//go:embed sql_role.json
var sqlRoleData []byte

func LoadSQLConfig() error {
	return csql.LoadSQLConfigData(sqlRoleData, "role/sql_role.json")
}

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
		Port: getEnv("PORT", "2427"),
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

	log.Println("Initializing Postgres connection for role module...")

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

	log.Println("Postgres connected successfully for role module")

	return nil
}

//go:embed api_role.json
var apiRoleData []byte

func getHandlerMap() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GetRoleListHandler":   GetRoleListHandler,
		"GetRoleHandler":       GetRoleHandler,
		"CreateRoleHandler":    CreateRoleHandler,
		"UpdateRoleHandler":    UpdateRoleHandler,
		"DeleteRoleHandler":   DeleteRoleHandler,
		"GetRoleByNameHandler": GetRoleByNameHandler,
	}
}

func RegisterRoutes(router *gin.Engine) {
	if err := api.RegisterRoutesFromJSONData(router, apiRoleData, "role/api_role.json", getHandlerMap()); err != nil {
		log.Fatalf("Failed to register role routes: %v", err)
	}
}

func Init() error {
	initConfig()
	if err := initDB(); err != nil {
		return err
	}
	return nil
}

func Close() {
	if pgDB != nil {
		pgDB.Close()
	}
}
