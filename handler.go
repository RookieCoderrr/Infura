package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)



func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w,"Hello world")
}

func AuthProjectId(w http.ResponseWriter, r *http.Request) {
	//获取URL中 apikey 参数
	params := mux.Vars(r)
	apikey:=params["id"]
	fmt.Println(apikey)
	//fmt.Println(strconv.Quote(r.Host))

	//连接数据库
	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co,_:=intializeMongoOnlineClient(cfg, ctx)

	//查看数据表中是否存在apikey
	filter:= bson.M{"apikey":apikey}
	var result *mongo.SingleResult
	result=co.Database("testdb").Collection("projects").FindOne(ctx,filter)

	//不存在apikey,直接返回http响应
	if result.Err() != nil {
		fmt.Println("=================PROJECT ID DOESN'T EXIST===============")
		//msg, _ :=json.Marshal(appError{result.Err(),"projectId "+projectId+" doesn't exist",8})
		//w.Header().Set("Content-Type","application/json")
		//w.Write(msg)
		fmt.Fprintf(w,"invalid projectId "+apikey)
		return
	} else {
		res,err:=result.DecodeBytes()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
		secretrequired := res.Lookup("secretrequired").Boolean()
		apisecret := res.Lookup("apisecret").String()
		if secretrequired {
			fmt.Println(r.BasicAuth())
			_,pwd,active := r.BasicAuth()
			if !active {
				fmt.Println("=================PROJECT SECRET REQUEIRED===============")
				fmt.Fprintf(w,"Project Secret required ")
				return
			} else {
				 if apisecret != strconv.Quote(pwd) {
					 fmt.Println("=================PROJECT SECRET ERROR===============")
					 fmt.Fprintf(w,"Project Secret error ")
					 return
				 }
			}

		}
		isLimited := checkProjectLimit(res,co,ctx,filter,w)
		if isLimited {
			return
		}
		isHostLimted := checkHostLimit(res,r,w)

		if isHostLimted{
			return
		}
		request := repostRequest(w,r)
		recordApi(request,apikey,co,ctx)
		recordProjectLimit(apikey,co,ctx)





	}
	//解析查询的document，避免再次查数据库

	//limit:=res.Lookup("limitperday").AsInt64()
	//host := res.Lookup("host").String()
	//fmt.Println(host)
	//if limit <= 0 {
	//	fmt.Fprintf(w,"your usage is up to limit")
	//	return
	//} else if host != strconv.Quote("") &&host != strconv.Quote(r.Host) {
	//		fmt.Fprintf(w,"rejected due to project ID settings")
	//		return
	//}
	//提取secretrequired字段

}

func errProjectId(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w,"project ID is required")
}

//func signUp(w http.ResponseWriter, r *http.Request) {
//	reader, err := r.MultipartReader()
//	var userName =""
//	var password = ""
//	var createTime = time.Now().Unix()
//	if err != nil {
//		log.Fatal(err)
//	}
//	for {
//		part, err := reader.NextPart()
//		if err == io.EOF {
//			break
//		}
//		data, _ := ioutil.ReadAll(part)
//		if part.FormName() == "Username" {
//			userName = string(data)
//		} else if part.FormName() == "Password"{
//			password = string(data)
//		} else {
//
//		}
//	}
//	cfg, err := OpenConfigFile()
//	if err != nil {
//		log.Fatal(" open file error")
//	}
//	ctx := context.TODO()
//	co,_:=intializeMongoOnlineClient(cfg, ctx)
//	filter:= bson.M{"email":userName}
//	var result *mongo.SingleResult
//	result=co.Database("infura").Collection("User").FindOne(ctx,filter)
//
//	if result.Err() != nil {
//		projectId := randomProjectId()
//		fmt.Println(userName,password,projectId,10000)
//		user := userInfo{ userName,password,projectId,10000, "",createTime}
//		var insertOne *mongo.InsertOneResult
//		insertOne, err = co.Database("infura").Collection("User").InsertOne(ctx,user)
//		fmt.Println("Connect to mainnet database")
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println("Inserted a  user in database",insertOne.InsertedID)
//		fmt.Fprintf(w,"Inserted a  user in database and your project Id is "+projectId)
//
//	} else {
//		fmt.Println("=================userName already exist===============")
//		//msg, _ :=json.Marshal(appError{result.Err(),"projectId "+projectId+" doesn't exist",8})
//		//w.Header().Set("Content-Type","application/json")
//		//w.Write(msg)
//		fmt.Fprintf(w,"User "+ userName+" has already signed up ")
//	}
//
//
//
//}

func createUser(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	var email = ""
	var password = ""
	if err != nil {
		log.Fatal(err)
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		data, _ := ioutil.ReadAll(part)
		if part.FormName() == "Email" {
			email = string(data)
		} else if part.FormName() == "Password" {
			password = string(data)
		} else {

		}
	}
	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co, _ := intializeMongoOnlineClient(cfg, ctx)
	filter := bson.M{"email": email}
	var result *mongo.SingleResult
	result = co.Database("infura").Collection("User").FindOne(ctx, filter)

	if result.Err() != nil {
		user := userInfo{email, password}
		var insertOne *mongo.InsertOneResult
		insertOne, err = co.Database("infura").Collection("User").InsertOne(ctx, user)
		fmt.Println("Connect to mainnet database")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Inserted a  user in database", insertOne.InsertedID)
		fmt.Fprintf(w,  "User "+email +" is created")
	} else {
		fmt.Println("=================userName already exist===============")
		//msg, _ :=json.Marshal(appError{result.Err(),"projectId "+projectId+" doesn't exist",8})
		//w.Header().Set("Content-Type","application/json")
		//w.Write(msg)
		fmt.Fprintf(w, "User "+email+" has already signed up ")

	}
}

func login(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	var email = ""
	var password = ""
	if err != nil {
		log.Fatal(err)
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		data, _ := ioutil.ReadAll(part)
		if part.FormName() == "Email" {
			email = string(data)
		} else if part.FormName() == "Password" {
			password = string(data)
		} else {

		}
	}
	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co, _ := intializeMongoOnlineClient(cfg, ctx)
	filter := bson.M{"email": email, "password": password}
	var result *mongo.SingleResult
	result = co.Database("infura").Collection("User").FindOne(ctx, filter)

	if result.Err() != nil {
		fmt.Println("===========Login failed ==========")
		fmt.Fprintf(w,"Email and password doesn't match.")
	} else {
		fmt.Println("===========Login success ==========")
		fmt.Fprintf(w,"Login success.")
	}


}

func createProject(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	var name =""
	var email = ""
	var createTime = time.Now().Unix()
	if err != nil {
		log.Fatal(err)
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		data, _ := ioutil.ReadAll(part)
		if part.FormName() == "Name" {
			name = string(data)
		} else if part.FormName() == "Email" {
			email = string(data)
		}
	}
	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co,_:=intializeMongoOnlineClient(cfg, ctx)
	filter:= bson.M{"email":email}
	var result *mongo.SingleResult
	result=co.Database("infura").Collection("User").FindOne(ctx,filter)
	if result.Err() != nil {
		fmt.Println("=================User doesn't exist===============")
		//msg, _ :=json.Marshal(appError{result.Err(),"projectId "+projectId+" doesn't exist",8})
		//w.Header().Set("Content-Type","application/json")
		//w.Write(msg)
		fmt.Fprintf(w,"User "+ email+" doesn't exist ")
	} else {
		filter:= bson.M{"email":email,"name":name}
		var result *mongo.SingleResult
		result=co.Database("infura").Collection("Project").FindOne(ctx,filter)

		if result.Err() != nil {
			projectId := randomProjectId()
			fmt.Println(name,projectId)
			project := projectInfo{email,name,projectId,10000,100,"",createTime}
			var insertOne *mongo.InsertOneResult
			insertOne, err = co.Database("infura").Collection("Project").InsertOne(ctx,project)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Created a project successfully",insertOne.InsertedID)
			fmt.Fprintf(w,"Create a project successfully, project Name: "+name+" projectId: "+projectId)


		}else {
			fmt.Println("=================Project name already exists===============")
			//msg, _ :=json.Marshal(appError{result.Err(),"projectId "+projectId+" doesn't exist",8})
			//w.Header().Set("Content-Type","application/json")
			//w.Write(msg)
			fmt.Fprintf(w,"Project "+ name+" already exists ")
		}
	}



}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	var name = ""
	var email = ""
	if err != nil {
		log.Fatal(err)
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		data, _ := ioutil.ReadAll(part)
		if part.FormName() == "Name" {
			name = string(data)
		} else if part.FormName() == "Email" {
			email = string(data)
		}
	}
	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal("open file error")
	}
	ctx := context.TODO()
	co, _ := intializeMongoOnlineClient(cfg, ctx)
	filter := bson.M{"email" : email, "name" : name}
	var result *mongo.SingleResult
	result = co.Database("infura").Collection("Project").FindOne(ctx,filter)

	if result.Err() != nil {
		fmt.Println("=================Delete project error ===============")
		fmt.Fprintf(w,"Project "+ name+" Email " +email + " doesn't exsit")

	} else {
		_, err = co.Database("infura").Collection("Project").DeleteOne(ctx, filter)
		fmt.Println("=================Delete project success ===============")
		fmt.Fprintf(w,"Project " + name + " Email " + "delete successfullly")
	}




}

func changeProjectName(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	var oldName = ""
	var newName = ""
	var email = ""
	if err != nil {
		log.Fatal(err)
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		data, _ := ioutil.ReadAll(part)
		if part.FormName() == "OldName" {
			oldName = string(data)
		} else if part.FormName() == "Email" {
			email = string(data)
		} else if part.FormName() == "NewName" {
			newName = string(data)
		}
	}
	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal("open file error")
	}
	ctx := context.TODO()
	co, _ := intializeMongoOnlineClient(cfg, ctx)
	filter := bson.M{"email" : email, "name" : oldName}
	var result *mongo.SingleResult
	result = co.Database("infura").Collection("Project").FindOne(ctx,filter)

	if result.Err() != nil {
		fmt.Println("=================Change project name error ===============")
		fmt.Fprintf(w,"Project "+ oldName+" Email " +email + " doesn't exsit")
	} else {
		update := bson.M{"$set": bson.M{"name":newName}}
		_, err = co.Database("infura").Collection("Project").UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("============= Project name updated==========")
		fmt.Fprintf(w,"Project "+oldName +" change to "+ newName)
	}

}

func addHost(w http.ResponseWriter, r *http.Request){
	reader, err := r.MultipartReader()
	var host string
	var email string
	var name string
	if err != nil {
		log.Fatal(err)
	}
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		data, _ := ioutil.ReadAll(part)
		if part.FormName() == "Host" {
			host = string(data)
		} else if part.FormName() == "Email" {
			email = string(data)
		} else if part.FormName() == "Name" {
			name = string(data)
		}
	}

	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co,_:=intializeMongoOnlineClient(cfg, ctx)
	filter:= bson.M{"email":email, "name":name}
	var result *mongo.SingleResult
	result = co.Database("infura").Collection("Project").FindOne(ctx,filter)
	if result.Err() != nil {
		fmt.Println("=================Add host error ===============")
		fmt.Fprintf(w,"Project "+ name+" Email " +email + " doesn't exsit")
	} else {
		update:=bson.M{"$set":bson.M{"host":host}}
		_, err = co.Database("infura").Collection("Project").UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("=================Host added===============")
		//msg, _ :=json.Marshal(appError{result.Err(),"projectId "+projectId+" doesn't exist",8})
		//w.Header().Set("Content-Type","application/json")
		//w.Write(msg)
		fmt.Fprintf(w,"Host "+ host+" is added ")
	}


}

func deleteHost(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	var name string
	var email string
	if err != nil {
		log.Fatal(err)
	}
	for {
		part,err := reader.NextPart()
		if err == io.EOF {
			break
		}
		data, _ :=ioutil.ReadAll(part)
		if part.FormName() == "Name" {
			name = string(data)
		} else if part.FormName() == "Email" {
			email = string(data)
		}
	}
	cfg, err := OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co,_:=intializeMongoOnlineClient(cfg, ctx)
	filter:= bson.M{"email":email, "name":name}
	var result *mongo.SingleResult
	result = co.Database("infura").Collection("Project").FindOne(ctx,filter)
	if result.Err() != nil {
		fmt.Println("=================Delete host error ===============")
		fmt.Fprintf(w,"Project "+ name+" Email " +email + " doesn't exsit")
	} else {
		update := bson.M{"$set":bson.M{"host":""}}
		_, err = co.Database("infura").Collection("Project").UpdateOne(ctx,filter,update)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("=================Host deleted===============")
		//msg, _ :=json.Marshal(appError{result.Err(),"projectId "+projectId+" doesn't exist",8})
		//w.Header().Set("Content-Type","application/json")
		//w.Write(msg)
		fmt.Fprintf(w,"User "+email+" Poject "+ name+" Host is deleted ")
	}


}

