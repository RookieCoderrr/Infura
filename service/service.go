package service

import (
	"Infura/tool"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type Service struct {
	Redis *redis.Client
	Db  *mongo.Client
}



func (s *Service)AuthProjectId(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	apikey:=params["id"]


	filter:= bson.M{"apikey":apikey}
	var result map[string]interface{}
	err :=s.Db.Database("testdb").Collection("projects").FindOne(context.TODO(),filter).Decode(&result)
	if err == mongo.ErrNoDocuments || err != nil {
		fmt.Println("=================PROJECT ID DOESN'T EXIST===============")
		fmt.Fprintf(w, "invalid projectId "+apikey)
		return
	}
	secretIdRequired := result["secretrequired"].(string)
	apiSecret := result["apisecret"].(string)
	if secretIdRequired == "true" {
		token := r.Header.Get("Token")
		timeStamp :=  r.Header.Get("TimeStamp")
		if token ==""  {
			fmt.Println("=================TOKEN NOT SET===============")
			fmt.Fprintf(w, "TOKEN NOT SET ")
			return
		} else if  timeStamp == "" {
			fmt.Println("=================TimeStamp NOT SET===============")
			fmt.Fprintf(w, "TIMESTAMP NOT SET ")
			return
		}  else {
			fmt.Println(token)
			fmt.Println(timeStamp)
			tool.EncodeMd5(token,apiSecret,timeStamp)
		}
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
	//isLimited := tool.CheckProjectLimit(res,s.Db,context.TODO(),filter,w)
	//if isLimited {
	//	return
	//}
	//isHostLimted := tool.CheckHostLimit(res,r,w)
	//
	//if isHostLimted{
	//	return
	//}
	//request := tool.RepostRequest(w,r)
	//tool.RecordApi(request,apikey,s.Db,context.TODO())
	//tool.RecordProjectLimit(apikey,s.Db,context.TODO())


}

func (s *Service)ErrProjectId(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w,"project ID is required")
}


