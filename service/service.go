package service

import (
	"Infura/tool"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/time/rate"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Service struct {
	Redis *redis.Client
	Db  *mongo.Client
	DbName string
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Change the the map to hold values of the type visitor.
var visitors = make(map[string]*visitor)
var mu sync.Mutex

// Run a background goroutine to remove old entries from the visitors map.

var (
	secretIdRequired bool
	apikey string
	apiSecret string
	host string
	request int32
	limitPerDay int32
	limitPerSecond int32
	origins primitive.A
	contractAddress primitive.A
	apiRequest primitive.A
	tokenClient string
	timeStampStr string
	timeStamp int64
)
func getVisitor(apiKey string, limitPerSecond int32) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[apiKey]
	if !exists {
		limiter := rate.NewLimiter(rate.Limit(limitPerSecond), 1)
		// Include the current time when creating a new visitor.
		visitors[apikey] = &visitor{limiter, time.Now()}
		return limiter
	}

	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	fmt.Println(v.limiter)
	return v.limiter
}
func CleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for apiKey, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, apiKey)
				fmt.Println("Delete Successfully")
			}
		}
		mu.Unlock()
	}
}

func (s *Service)AuthProjectId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	apikey =params["id"]
	host = r.Host
	fmt.Println(host)

	filter:= bson.M{"apikey":apikey}
	var result map[string]interface{}
	err := s.Db.Database(s.DbName).Collection("projects").FindOne(context.TODO(),filter).Decode(&result)
	if err == mongo.ErrNoDocuments || err != nil {
		fmt.Println("=================PROJECT ID DOESN'T EXIST===============")
		fmt.Fprintf(w, "invalid projectId "+apikey)
		return
	}
	limitPerSecond = result["limitpersecond"].(int32)
	secretIdRequired = result["secretrequired"].(bool)
	apiSecret = result["apisecret"].(string)
	request = result["request"].(int32)
	limitPerDay = result["limitperday"].(int32)
	origins = result["origin"].(primitive.A)
	contractAddress = result["contractAddress"].(primitive.A)
	apiRequest = result["ApiRequest"].(primitive.A)
	tokenClient = r.Header.Get("Token")
	timeStampStr =  r.Header.Get("TimeStamp")
	timeStamp, err = strconv.ParseInt(timeStampStr, 10, 64)

	limiter := getVisitor(apikey,limitPerSecond)
	if limiter.Allow() == false {
		http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
		return
	}
	if secretIdRequired  {
		if !tool.CheckHostLimit(origins,host) {
			fmt.Println("=================Host not permitted===============")
			fmt.Fprintf(w, "Host not permitted.")
			return
		}
		if !tool.CheckProjectLimit(limitPerDay,request) {
			fmt.Println("=================Reach Daily Limit===============")
			fmt.Fprintf(w, "This projectId has reached the daily limit.")
			return
		}

		if tokenClient ==""   {
			fmt.Println("=================TOKEN NOT SET===============")
			fmt.Fprintf(w, "Token not set in http header")
			return
		} else if  timeStampStr == "" || err != nil{
			fmt.Println("=================TimeStamp NOT SET===============")
			fmt.Fprintf(w, "TimeStamp not set in http header")
			return
		} else if len(timeStampStr)!= 13  {
			fmt.Println("=================TimeStamp NOT Standard===============")
			fmt.Fprintf(w, "TimeStamp not standard")
			return
		} else if  time.Now().UnixNano()/ 1000000 - timeStamp >= 3600000 {
			fmt.Println("=================TimeStamp HAS EXPIRED===============")
			fmt.Fprintf(w, "TimeStamp has expired")
			return
		} else {
			//fmt.Println(tokenClient)
			//fmt.Println(timeStamp)
			tokenServer := tool.EncodeMd5(apikey,apiSecret,timeStampStr)
			if tokenServer == tokenClient {
				req := tool.RepostRequest(w,r,apiRequest,contractAddress)
				if req != nil {
					tool.RecordApi(req,apikey,s.Db,context.TODO(),s.DbName)
					tool.RecordRequest(apikey,s.Db,context.TODO(),s.DbName)
				}
				return
			} else {
				fmt.Println("=================TOKEN INVALID===============")
				fmt.Fprintf(w, "Token invalid")
				return
			}
		}
	} else {
		if !tool.CheckHostLimit(origins,host) {
			fmt.Println("=================Host not permitted===============")
			fmt.Fprintf(w, "Host not permitted.")
			return
		}
		if !tool.CheckProjectLimit(limitPerDay,request) {
			fmt.Println("=================Reach Daily Limit===============")
			fmt.Fprintf(w, "This projectId has reached the daily limit.")
			return
		}

		req := tool.RepostRequest(w,r,apiRequest,contractAddress)
		if req != nil {
			tool.RecordApi(req,apikey,s.Db,context.TODO(),s.DbName)
			tool.RecordRequest(apikey,s.Db,context.TODO(),s.DbName)
		}
		return
	}

	//tool.RecordProjectLimit(apikey,s.Db,context.TODO())
	//fmt.Println(r.BasicAuth())
	//_,pwd,active := r.BasicAuth()
	//if !active {
	//	fmt.Println("=================PROJECT SECRET REQUEIRED===============")
	//	fmt.Fprintf(w,"Project Secret required ")
	//	return
	//} else {
	//	if apiSecret != strconv.Quote(pwd) {
	//		fmt.Println("=================PROJECT SECRET ERROR===============")
	//		fmt.Fprintf(w,"Project Secret error ")
	//		return
	//	}
	//}


}

func (s *Service)ErrProjectId(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w,"project ID is required")
}

