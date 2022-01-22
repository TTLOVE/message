package data

import (
	"fmt"
	"log"
	"message/internal/conf"
	"time"

	"github.com/Shopify/sarama"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"github.com/jinzhu/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewMessageRepo)

// Data .
type Data struct {
	*gorm.DB
	KafkaClient sarama.SyncProducer
}

// NewData .
func NewData(c *conf.ConfigData) (*Data, func(), error) {
	cleanup := func() {
		fmt.Println("closing the data resources")
	}

	db, err := handleDB(c)
	if err != nil {
		log.Fatalln("db connect fail:", err)
	}
	kafkaClient, err := handleKafka(c)
	if err != nil {
		log.Fatalln("new kafka fail:", err)
	}

	return &Data{
		DB:          db,
		KafkaClient: kafkaClient,
	}, cleanup, nil
}

func handleDB(c *conf.ConfigData) (*gorm.DB, error) {
	db, err := gorm.Open(c.Database.Driver, c.Database.Dsn)
	if err != nil {
		fmt.Println("connet fail:", err)
		return nil, err
	}
	db.DB().SetMaxOpenConns(0)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetConnMaxLifetime(time.Hour * 1)
	db.Debug()
	return db, nil
}

func handleKafka(c *conf.ConfigData) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // 发送完数据需要leader和follow都确认
	config.Producer.Partitioner = sarama.NewRandomPartitioner // 新选出一个partition
	config.Producer.Return.Successes = true                   // 成功交付的消息将在success channel返回

	// 连接kafka
	KafkaClient, err := sarama.NewSyncProducer([]string{c.GetKafka().Addr}, config)
	if err != nil {
		fmt.Println("producer closed, err:", err)
		return nil, err
	}

	return KafkaClient, nil
}
