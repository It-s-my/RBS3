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

// Config структура для хранения порта
type Config struct {
	Port int `json:"port"`
}

// HandleFileSort - обрабатывает HTTP запросы на сервере.
func HandleFileSort(w http.ResponseWriter, r *http.Request) {

	// Получаем значения параметров "root" и "sort" из URL запроса
	root := r.URL.Query().Get("root")
	sortM := r.URL.Query().Get("sort")

	// Вызываем функцию Sortfile из пакета syst для сортировки файлов
	data := syst.Sortfile(root, sortM)

	// Преобразуем данные в формат JSON
	resp, err := json.Marshal(data)
	// Если произошла ошибка при маршалинге данных, логируем ошибку и отправляем HTTP ошибку
	if err != nil {
		log.Printf("%v %v", r.URL, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовок Content-Type как application/json
	w.Header().Set("Content-Type", "application/json")
	// Устанавливаем статус код HTTP ответа на 200 OK
	w.WriteHeader(http.StatusOK)
	// Отправляем ответ клиенту
	w.Write(resp)
}

// main - точка входа в программу
func main() {
	// Читаем конфигурацию из файла
	file, err := os.Open("config/config.json")
	if err != nil {
		fmt.Println("Ошибка открытия файла", err)
		return
	}
	defer file.Close()

	// config - переменная для хранения конфигурации
	var config Config
	// Декодируем данные из файла в структуру Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Println("Ошибка декодирования данных:", err)
		return
	}
	// Создаем HTTP сервер с настройками из конфигурации
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: http.HandlerFunc(HandleFileSort),
	}
	// Запускаем HTTP сервер асинхронно в горутине
	go func() {
		fmt.Println("Пытаюсь запустить сервер на порту", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Не удалось запустить сервер на порту %d: %v\n", config.Port, err)
			return
		}
	}()
	// Создаем канал stop для получения сигналов остановки сервера
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Ждем сигнала остановки
	<-stop

	fmt.Println("\nОстанавливаю сервер...")

	// Создаем контекст с таймаутом для остановки сервера
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Останавливаем сервер с учетом контекста
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Ошибка остановки сервера:", err)
	}

	fmt.Println("Сервер остановлен корректно.")
}
