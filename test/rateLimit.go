package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main () {
	url := "http://127.0.0.1:1926/projectId/465355e80ce88bbf542a58eee1dadedf"
	client := &http.Client{}
	data := make(map[string]interface{})
	data["name"] = "2.0"
	data["method"] = "GetAssetInfoByContractHash"
	data["params"] = map[string]string{"ContractHash":"0xcd10d9f697230b04d9ebb8594a1ffe18fa95d9ad"}
	bytesData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST",url,bytes.NewReader(bytesData))

	for i := 0; i < 100; i++ {
		resp, _ := client.Do(req)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))

	}
}