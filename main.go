package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)


func main()  {
	fmt.Println("Server start")
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/projectId/{id}",AuthProjectId)
	muxRouter.HandleFunc("/projectId/",errProjectId)
	muxRouter.HandleFunc("/createProject",createProject)
	muxRouter.HandleFunc("/login",login)
	muxRouter.HandleFunc("/createUser",createUser)
	muxRouter.HandleFunc("/changeProjectName",changeProjectName)
	muxRouter.HandleFunc("/deleteProject",deleteProject)
	muxRouter.HandleFunc("/addHost",addHost)
	muxRouter.HandleFunc("/deleteHost",deleteHost)
	//muxRouter.HandleFunc("/signUp",signUp)
	muxRouter.HandleFunc("/{params}",errProjectId)
	muxRouter.HandleFunc("/",errProjectId)

	//muxRouter.HandleFunc("/{name}/{country}",ShowVisitorInfo)
	http.ListenAndServe(":1926",muxRouter)
	//RegisterRouters(muxRouter)
	//server := &http.Server{
	//	Addr: ":1926",
	//	Handler: muxRouter,
	//}
	//err := server.ListenAndServe()
	//if err != nil {
	//	log.Fatal(err)
	//}
}


