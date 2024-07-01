package main

import (
	"Server/syst"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Port int `json:"port"`
}

// HandleFileSort - обрабатывает HTTP запросы на сервере.
func HandleFileSort(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("root")
	sortM := r.URL.Query().Get("sort")

	data := syst.Sortfile(root, sortM)

	resp, err := json.Marshal(data)
	if err != nil {
		log.Printf("%v %v", r.URL, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	w.Write(resp)
}

func main() {
	// Читаем конфигурацию из файла
	file, err := os.Open("config/config.json")
	if err != nil {
		fmt.Println("Ошибка открытия файла", err)
		return
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Println("Ошибка декодирования данных:", err)
		return
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(HandleFileSort),
	}

	go func() {
		fmt.Println("Пытаюсь запустить сервер на порту", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Не удалось запустить сервер на порту %d: %v\n", config.Port, err)
			return
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Ждем сигнала остановки
	<-stop

	fmt.Println("\nОстанавливаю сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Ошибка остановки сервера:", err)
	}

	fmt.Println("Сервер остановлен корректно.")
}
