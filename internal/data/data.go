package data

import (
	"fmt"
	"message/internal/conf"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/jinzhu/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewMessageRepo)

// Data .
type Data struct {
	*gorm.DB
}

// NewData .
func NewData(c *conf.ConfigData) (*Data, func(), error) {
	cleanup := func() {
		fmt.Println("closing the data resources")
	}

	db, err := gorm.Open(c.Database.Driver, c.Database.Dsn)
	if err != nil {
		fmt.Println("connet fail:", err)
	}
	db.DB().SetMaxOpenConns(0)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetConnMaxLifetime(time.Hour * 1)
	db.Debug()

	return &Data{DB: db}, cleanup, nil
}
