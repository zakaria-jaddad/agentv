package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/zakaria-jaddad/agentv/internal/agentv"
	"github.com/zakaria-jaddad/agentv/internal/config"
)

type Manager struct {
	Agent  *agentv.Agentv
	Config *config.Config
}

func main() {

	var confpath string
	flag.StringVar(&confpath, "config", "./agentv.yml", "Configuration file path")
	flag.Parse()

	conf, err := config.Load(confpath)
	if err != nil {
		log.Fatal(err)
	}

	if err := conf.Validate(); err != nil {
		log.Fatal(err)
	}

	agentv := agentv.New(conf.Agent.Name)
	agentv.DiscoverSystemInfo()

	manager := Manager{Agent: agentv, Config: conf}

	// information
	// hostname, err := os.Hostname()
	// if err != nil {
	// 	log.Fatal("ERROR: ", err)
	// }
	// platform := runtime.GOOS
	// arch := runtime.GOARCH
	// // "agentVersion": "1.0.0",
	//
	// payload := map[string]string{
	// 	"hostname":     hostname,
	// 	"token":        token,
	// 	"platform":     platform,
	// 	"architecture": arch,
	// }
	// jsonData, _ := json.Marshal(payload)
	//
	// req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	// if err != nil {
	// 	log.Fatal("ERROR: ", err)
	// }
	// req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("User-Agent", "ASSAS/V1")
	//
	// // Client Creating
	// client := &http.Client{
	// 	Timeout: 10 * time.Second,
	// }
	//
	// res, err := client.Do(req)
	// if err != nil {
	// 	log.Fatal("ERROR: ", err)
	// }
	// defer res.Body.Close()
	//
	// // Response
	// body, _ := io.ReadAll(res.Body)
	// fmt.Println("Status:", res.Status)
	// fmt.Println("Body:", string(body))
}
