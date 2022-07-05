package main

import (
	"Infura/service"
	"Infura/tool"
	"context"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/rs/cors"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func init() {
	go service.CleanupVisitors()
}

func main()  {
	fmt.Println("Server start")
	//fmt.Println(time.Now().UnixNano()/ 1000000)
	//fmt.Println( time.Now().UnixNano()/ 1000000 - 1647398951614)
	//service.Sub(1647430804000,1647420004000)
	rt := os.ExpandEnv("${RUNTIME}")
	switch rt {
	case "test":
		fmt.Println("Runtime: test")
	case "staging":
		fmt.Println("Runtime: staging")
	case "testmagnet":
		fmt.Println("Runtime: testmagnet")
	default:
		fmt.Println("Runtime: default")
	}
	whiteList := map[string]int{
		"a": 1,
	}
	tool.EncodeMd5("d74fd1c42f4bc21114d0c5f1500f366b","80e8365ede8806b5daf0d72f62c01e22","1651724951696")
	cfg, err :=  tool.OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co,dbName:=tool.InitializeMongoOnlineClient(cfg, ctx)
	s := &service.Service{
		Db: co,
		DbName: dbName,
		WhiteList: whiteList,
	}
	r := mux.NewRouter()
	r.HandleFunc("/projectId/{id}",s.AuthProjectId)
	r.HandleFunc("/projectId/",s.ErrProjectId)
	r.HandleFunc("/{params}",s.ErrProjectId)
	r.HandleFunc("/",s.ErrProjectId)
	c := cron.New()
	c.AddFunc("@daily",func(){
		fmt.Println("Start daily job")
		tool.ResetRequestCount(co,context.TODO(),dbName)
	})
	c.Start()
	handler := cors.Default().Handler(r)
	http.ListenAndServe(":1926",handler)

}


