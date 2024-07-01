package main

import (
	"Server/syst" // Импортируем пакет
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Config struct {
	Port int `json:"port"`
}

func main() {
	// Читаем конфигурацию из файла
	configData, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		fmt.Println("Error unmarshalling config data:", err)
		return
	}

	http.HandleFunc("/path", syst.HandleFileSort)

	fmt.Println("Server is running on port", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
