package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Sg struct {
	Id   string `json:"id"`
	Port int64  `json:"port"`
	Comment string `json:"comment"`
}

type Config struct {
	Sgs []Sg `json:"sgs"`
}

type IpifyResponse struct {
	Ip string `json:"ip"`
}

const IPApi string = "https://api.ipify.org?format=json"
const DefaultConfigFile string = "~/.sgsync/config.json"

func main() {
	var configFile string

	if len(os.Args) > 1 {
		configFile = os.Args[1]
	} else {
		configFile = DefaultConfigFile
	}

	sgs := initApp(configFile)
	myIp := getMyIp()
	svc := initAws()

	// TODO: run a daemon
	syncSgIps(myIp, svc, sgs)
}

func getMyIp() string {
	resp, err := http.Get(IPApi)

	if err != nil {
		fmt.Println(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var ipifyResp IpifyResponse
	json.Unmarshal(body, &ipifyResp)
	return ipifyResp.Ip
}

func initApp(configFile string) []Sg {
	jsonFile, err := os.Open(configFile)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	var ids Config

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &ids)
	return ids.Sgs
}
