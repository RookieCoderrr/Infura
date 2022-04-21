package main

import (
	"Infura/service"
	"Infura/tool"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
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
	default:
		fmt.Println("Runtime: default")
	}
	tool.EncodeMd5("465355e80ce88bbf542a58eee1dadedf","4ab781f72c0bdb4edfe1565282ff93e0","1648698459000")
	cfg, err :=  tool.OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co,dbName:=tool.InitializeMongoOnlineClient(cfg, ctx)
	s := &service.Service{
		Db: co,
		DbName: dbName,
	}
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/projectId/{id}",s.AuthProjectId)
	muxRouter.HandleFunc("/projectId/",s.ErrProjectId)
	muxRouter.HandleFunc("/{params}",s.ErrProjectId)
	muxRouter.HandleFunc("/",s.ErrProjectId)
	c := cron.New()
	c.AddFunc("@daily",func(){
		fmt.Println("Start daily job")
		tool.ResetRequestCount(co,context.TODO(),dbName)
	})
	c.Start()
	http.ListenAndServe(":1926",muxRouter)

}


