package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Database_main struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
		Database string `yaml:"database"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database_main"`
	Database_test struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Pass     string `yaml:"pass"`
		Database string `yaml:"database"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database_test"`
}

//type userInfo struct {
//	Email string
//	Password string
//	ProjectId string
//	Limit int
//	Host string
//	CreateTime int64
//}

type userInfo struct {
	Email string
	Password string
}
type projectInfo struct {
	Email string
	Name string
	ProjectId string
	LimitPerDay int
	LimitPerSecond int
	Host string
	CreateTime int64
}
type rpcInfo struct {
	ProjectId string
	Method string
	Timestamp int64
}
func randomProjectId() string{
	rand.Seed(time.Now().UnixNano())
	chars :=[]rune("abcdefghijklmnopqrstuvwxyz" + "0123456789")
	length := 30
	var b strings.Builder
	for i := 0 ; i <length; i ++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func intializeMongoOnlineClient(cfg Config, ctx context.Context) (*mongo.Client, string) {
	rt := os.ExpandEnv("${RUNTIME}")
	var clientOptions *options.ClientOptions
	var dbOnline string
	if rt != "mainnet" && rt !="testnet"{
		rt = "mainnet"
	}
	switch rt {
	case "mainnet":
		clientOptions = options.Client().ApplyURI("mongodb://" + cfg.Database_main.User + ":" + cfg.Database_main.Pass + "@" + cfg.Database_main.Host + ":" + cfg.Database_main.Port + "/" + cfg.Database_main.Database)
		dbOnline = cfg.Database_main.Database
	case "testnet":
		clientOptions = options.Client().ApplyURI("mongodb://" + cfg.Database_test.User + ":" + cfg.Database_test.Pass + "@" + cfg.Database_test.Host + ":" + cfg.Database_test.Port + "/" + cfg.Database_test.Database)
		dbOnline = cfg.Database_test.Database
	}


	clientOptions.SetMaxPoolSize(50)
	co, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("momgo connect error")
	}
	err = co.Ping(ctx, nil)
	if err != nil {
		log.Fatal("ping mongo error")
	}
	fmt.Println("Connect mongodb success")
	return co, dbOnline
}
func OpenConfigFile() (Config, error) {
	absPath, _ := filepath.Abs("config.yml")
	f, err := os.Open(absPath)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, err
}

