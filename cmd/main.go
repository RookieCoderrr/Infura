package main

import (
	"Infura/service"
	"Infura/tool"
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)


func main()  {
	fmt.Println("Server start")
	//service.Sub(1644508800000,1644422400000)
	tool.EncodeMd5("VZTxpIoXsrOKhFbF","nXlGyreGpwmmgclH","1644422400000")
	cfg, err :=  tool.OpenConfigFile()
	if err != nil {
		log.Fatal(" open file error")
	}
	ctx := context.TODO()
	co,_:=tool.IntializeMongoOnlineClient(cfg, ctx)
	s := &service.Service{
		Db: co,
	}
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/projectId/{id}",s.AuthProjectId)
	muxRouter.HandleFunc("/projectId/",s.ErrProjectId)
	muxRouter.HandleFunc("/{params}",s.ErrProjectId)
	muxRouter.HandleFunc("/",s.ErrProjectId)
	//c := cron.New()
	//c.AddFunc("@daily",func(){
	//	fmt.Println("Start hourly job")
	//	resetMethodCount()
	//})
	//c.Start()
	http.ListenAndServe(":1926",muxRouter)

}


