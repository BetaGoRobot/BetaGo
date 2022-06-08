package dbpack

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Administrator is the struct of administrator
type Administrator struct {
	gorm.Model
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	Level    int64  `json:"level"`
}

var (
	isTest = os.Getenv("IS_TEST")
)

func init() {
	// try get db conn
	if GetDbConnection() == nil {
		log.Println("get db connection error")
		os.Exit(-1)
	}
}

// GetDbConnection  returns the db connection
//  @return *gorm.DB
func GetDbConnection() *gorm.DB {
	var dsn string
	if isTest == "true" {
		dsn = "host=localhost user=postgres password=heyuheng1.22.3 dbname=betago port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	} else {
		dsn = "host=betago-pg user=postgres password=heyuheng1.22.3 dbname=betago port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "betago.",
			SingularTable: false,
		},
	})
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return db
}
