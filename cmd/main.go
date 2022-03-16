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
)


func main()  {
	fmt.Println("Server start")
	//fmt.Println(time.Now().UnixNano()/ 1000000)
	//fmt.Println( time.Now().UnixNano()/ 1000000 - 1647398951614)
	//service.Sub(1647430804000,1647420004000)
	tool.EncodeMd5("vnQiyDzZKufyyrQw","pPozWsLfNjQRQhnV","1647409947700")
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


