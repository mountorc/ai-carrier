package skill

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

//go:embed sql_skill.json
var sqlSkillData []byte

func LoadSQLConfig() error {
	return csql.LoadSQLConfigData(sqlSkillData, "skill/sql_skill.json")
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

	log.Println("Initializing Postgres connection for skill module...")

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

	log.Println("Postgres connected successfully for skill module")

	return nil
}

//go:embed api_skill.json
var apiSkillData []byte

func getHandlerMap() map[string]gin.HandlerFunc {
	return map[string]gin.HandlerFunc{
		"GetSkillListHandler":          GetSkillListHandler,
		"getSkillHandler":              getSkillHandler,
		"CreateSkillHandler":           CreateSkillHandler,
		"RegisterSkillHandler":         RegisterSkillHandler,
		"UploadSkillHandler":           UploadSkillHandler,
		"UpdateSkillHandler":           UpdateSkillHandler,
		"DeleteSkillHandler":           DeleteSkillHandler,
		"ListSkillVersionsHandler":     ListSkillVersionsHandler,
		"GetSkillVersionHandler":       GetSkillVersionHandler,
		"GetLatestSkillVersionHandler": GetLatestSkillVersionHandler,
	}
}

func RegisterRoutes(router gin.IRouter) {
	if err := api.RegisterRoutesFromJSONData(router, apiSkillData, "skill/api_skill.json", getHandlerMap()); err != nil {
		log.Fatalf("Failed to register skill routes: %v", err)
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
