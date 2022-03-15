package tool

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

func IntializeMongoOnlineClient(cfg Config, ctx context.Context) (*mongo.Client, string) {
	var clientOptions *options.ClientOptions
	var dbOnline string
	clientOptions = options.Client().ApplyURI("mongodb://"  +cfg.Database_main.Host + ":" + cfg.Database_main.Port )
	dbOnline = cfg.Database_main.Database


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
func CheckProjectLimit (res bson.Raw,client *mongo.Client ,ctx context.Context, filter bson.M, w http.ResponseWriter) bool{
	limitPerDay:=res.Lookup("limitperday").AsInt64()

	var resultLimit *mongo.SingleResult
	resultLimit= client.Database("testdb").Collection("projectlimits").FindOne(ctx,filter)
	resLimit, err := resultLimit.DecodeBytes()
	var limit int64
	if err != nil {
		fmt.Println("No project limit recorded")
		limit = 0
		return false
	} else {
		fmt.Println(resLimit)
		limit = resLimit.Lookup("methodcount").AsInt64()
		if limit > limitPerDay {
			fmt.Fprintf(w,"your usage is up to limit")
			return true
		} else {
			return false
		}
	}
}
func CheckHostLimit (res bson.Raw,r *http.Request,w http.ResponseWriter) bool{
	hostList := res.Lookup("origin").Array()
	host := r.Host
	fmt.Println(host)
	for i := 0; i < 100; i++ {
		 hostData,err:=hostList.IndexErr(uint(i))

		 if i == 0 && err != nil {
		 	fmt.Println("===========noHost=============")
		 	return false
		 } else if hostData.Value().String() == strconv.Quote(host) {
			 fmt.Println("===========Hostverified=============")
		 	return false
		 } else if err != nil {
			 fmt.Fprintf(w,"Your host is limited")
			 fmt.Println("===========No match Host =============")
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

	if err != nil {
		http.Error(w, "can't decoding in JSON", http.StatusBadRequest)
	}
	requestBody := bytes.NewBuffer(body)
	w.Header().Set("Content-Type", "application/json")
	resp, err := http.Post("https://neofura.ngd.network", "application/json", requestBody)
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
func RecordApi  (req map[string]interface{},apikey string, client *mongo.Client ,ctx context.Context) {
	method := req["method"].(string)
	createTime := time.Now().Unix()
	rpc := rpcInfo{apikey,method,createTime}
	insertOne, err := client.Database("testdb").Collection("projectrpcrecords").InsertOne(ctx,rpc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a RPC method in database",insertOne)

}

func RecordProjectLimit (apikey string, client *mongo.Client ,ctx context.Context) {
	filter:= bson.M{"apikey":apikey}
	var result *mongo.SingleResult
	result=client.Database("testdb").Collection("projectlimits").FindOne(ctx,filter)
	if result.Err() != nil {
		createTime := time.Now().Unix()
		methodCount := projectLimit{apikey,0,createTime}
		insertOne, err := client.Database("testdb").Collection("projectlimits").InsertOne(ctx,methodCount)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Inserted a project limit in database",insertOne)

	} else {
		update:=bson.M{"$inc" :bson.M{"methodcount":1}}
		updateOne, err :=client.Database("testdb").Collection("projectlimits").UpdateOne(ctx,filter,update)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("update a project limit in database",updateOne)
	}


}
//func ResetMethodCount () {
//	cfg, err := OpenConfigFile()
//	if err != nil {
//		log.Fatal(" open file error")
//	}
//	ctx := context.TODO()
//	co,_:=intializeMongoOnlineClient(cfg, ctx)
//	update:=bson.M{"$set" :bson.M{"methodcount":0}}
//	updateMany, err := co.Database("testdb").Collection("projectlimits").UpdateMany(ctx,bson.M{},update)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println("update all project limits to 0 in database",updateMany)
//
//}
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

func EncodeMd5(secretId string, projectId string, timeStamp string) string {
	has := md5.New()
	has.Write([]byte(projectId+secretId+timeStamp))
	b := has.Sum(nil)
	md5 := hex.EncodeToString(b)
	fmt.Println(md5)
	return md5
}


