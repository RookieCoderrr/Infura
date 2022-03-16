package tool

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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


type rpcInfo struct {
	Apikey string
	Method string
	Timestamp int64
}
type projectLimit struct {
	Apikey string
	MethodCount int
	Timestamp int64
}

func InitializeMongoOnlineClient(cfg Config, ctx context.Context) (*mongo.Client, string) {
	rt := os.ExpandEnv("${RUNTIME}")
	var clientOptions *options.ClientOptions
	var dbOnline string
	switch rt {
	case "test":
		clientOptions = options.Client().ApplyURI("mongodb://"  +cfg.Database_test.Host + ":" + cfg.Database_test.Port )
		dbOnline = cfg.Database_main.Database
	case "staging":
		clientOptions = options.Client().ApplyURI("mongodb://"  +cfg.Database_main.Host + ":" + cfg.Database_main.Port )
		dbOnline = cfg.Database_main.Database
	default:
		clientOptions = options.Client().ApplyURI("mongodb://"  +cfg.Database_main.Host + ":" + cfg.Database_main.Port )
		dbOnline = cfg.Database_main.Database
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
func CheckProjectLimit (limitPerDay int32, request int32) bool{
	if limitPerDay >= request {
		return true
	} else {
		return false
	}
}
func CheckHostLimit (origins primitive.A, host string) bool{
	if len(origins) == 0 {
		return true
	}
	for i := 0; i < len(origins); i++ {
		if host  == origins[i] {
			return true
		}
	}
	return false

}
func RepostRequest(w http.ResponseWriter, r *http.Request) map[string]interface{}{
	body, err := ioutil.ReadAll(r.Body)
	request := make(map[string]interface{})
	err = json.Unmarshal(body, &request)
	fmt.Println(request)
	requestBody := bytes.NewBuffer(body)
	w.Header().Set("Content-Type", "application/json")
	rt := os.ExpandEnv("${RUNTIME}")
	var resp *http.Response
	switch rt {
	case "test":
		resp, err = http.Post("https://testneofura.ngd.network:444", "application/json", requestBody)
	case "staging":
		resp, err = http.Post("https://neofura.ngd.network", "application/json", requestBody)
	default:
		resp, err = http.Post("https://neofura.ngd.network", "application/json", requestBody)
	}
	if err != nil {
		fmt.Fprintf(w,"Repost error")
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(w,"Read err")
	}
	w.Write(body)
	return request
}
func RecordApi  (req map[string]interface{},apikey string, client *mongo.Client ,ctx context.Context,dbName string) {
	method := req["method"].(string)
	createTime := time.Now().UnixNano()/1000000
	rpc := rpcInfo{apikey,method,createTime}
	insertOne, err := client.Database(dbName).Collection("projectrpcrecords").InsertOne(ctx,rpc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a RPC method in database",insertOne)

}

func RecordRequest (apikey string, client *mongo.Client ,ctx context.Context, dbName string) {
	filter:= bson.M{"apikey":apikey}
	var result *mongo.SingleResult
	result=client.Database(dbName).Collection("projects").FindOne(ctx,filter)
	if result.Err() != nil {
		return

	} else {
		update:=bson.M{"$inc" :bson.M{"request":1}}
		updateOne, err :=client.Database(dbName).Collection("projects").UpdateOne(ctx,filter,update)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("update a project limit in database",updateOne)
	}


}
func ResetRequestCount (co *mongo.Client,ctx context.Context,dbName string) {

	update:=bson.M{"$set" :bson.M{"request":0}}
	updateMany, err := co.Database("testdb").Collection("projects").UpdateMany(ctx,bson.M{},update)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("update all project request to 0 in database",updateMany)

}
func OpenConfigFile() (Config, error) {
	absPath, _ := filepath.Abs("../config.yml")
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

func EncodeMd5( projectId string,secretId string, timeStamp string) string {
	has := md5.New()
	has.Write([]byte(projectId+secretId+timeStamp))
	b := has.Sum(nil)
	md5 := hex.EncodeToString(b)
	//fmt.Println(md5)
	return md5
}


