package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/nusabangkit/finex/types"
	"github.com/nusabangkit/pkg/services"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var DataBase *gorm.DB
var Logger *logrus.Entry
var KafkaProducer *services.KafkaProducer
var RangoClient *services.RangoClient
var Referral *types.Referral
var Redis *services.RedisClient

func InitializeConfig() error {
	Logger = services.NewLoggerService("Finex")
	db, err := NewDatabase()
	if err != nil {
		return err
	}

	DataBase = db
	KafkaProducer, err = services.NewKafkaProducer(strings.Split(os.Getenv("KAFKA_URL"), ","), Logger)
	if err != nil {
		return err
	}

	RangoClient, err = services.NewRangoClient(KafkaProducer)
	if err != nil {
		return err
	}

	Redis, err = services.NewRedisClient(os.Getenv("REDIS_URL"))
	Logger.Info(os.Getenv("REDIS_URL"))
	if err != nil {
		return err
	}

	if err := NewInfluxDB(); err != nil {
		return err
	}

	buf, err := ioutil.ReadFile("config/config.yaml")
	if err != nil {
		return err
	}

	var config *types.Config
	if yaml.Unmarshal(buf, &config) != nil {
		return err
	}

	Referral = config.Referral

	return nil
}
