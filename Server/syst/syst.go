package syst

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

//Объявляем константы, которые помогут переводить размер файлов из байтов.
const thousand float64 = 1000
const GB int64 = 1000 * 1000 * 1000
const MB int64 = 1000 * 1000
const KB int64 = 1000

//Структура, в которой будут записаны название, тип и размер файлов.
type FileInfo struct {
	Name string
	Type string
	Size int64
}

//sortBySizeAsc - принимает срез файлов для сортировки от меньшего к большему.
func sortBySizeAsc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size < files[j].Size
	})
}

//sortBySizeDesc - принимает срез файлов и элементы сортируются в порядке убывания размера.
func sortBySizeDesc(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})
}

//walk - принимает строку root которая содержит информацию о размере каждой директории, начиная с корневой директории root.
func walk(root string) (map[string]int64, error) {

	directorySizes := make(map[string]int64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error { //
		if err != nil {
			fmt.Printf("Ошибка доступа %q: %v\n", path, err)
			return err
		}

		if info.IsDir() {
			fmt.Printf("Проход директории: %s\n", path)
			wg.Add(1)
			go func(dirPath string) {
				defer wg.Done()
				var dirSize int64
				err := filepath.Walk(dirPath, func(subPath string, subInfo os.FileInfo, subErr error) error {
					if subErr != nil {
						fmt.Printf("Ошибка доступа %q: %v\n", subPath, subErr)
						return subErr
					}
					dirSize += subInfo.Size()
					return nil
				})
				if err != nil {
					fmt.Printf("Ошибка размера каталога %q: %v\n", dirPath, err)
				}
				mu.Lock()
				directorySizes[dirPath] = dirSize
				mu.Unlock()
			}(path)
		}

		return nil
	})
	wg.Wait()

	if walkErr != nil {
		return nil, walkErr
	}

	return directorySizes, nil
}

//HandleFileSort - обрабатывает HTTP запросы на сервере.
func HandleFileSort(w http.ResponseWriter, r *http.Request) {
	root := r.URL.Query().Get("root")
	sortM := r.URL.Query().Get("sort")

	data := Sort_file(root, sortM)

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

func Sort_file(rootPath string, sortOrder string) []FileInfo {

	start := time.Now()

	files, err := filepath.Glob(filepath.Join(rootPath, "*"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	File_Info := make([]FileInfo, 0)

	directories, walkErr := walk(rootPath)
	if walkErr != nil {
		fmt.Println("Ошибка при обходе файловой системы:", walkErr)
		os.Exit(1)
	}

	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			fmt.Println("Ошибка чтения информации о файле:", err)
			continue
		}

		fileType := "file"
		if fileInfo.IsDir() {
			fileType = "directory"
		}

		var fileSize int64
		if fileInfo.IsDir() {
			fileSize = directories[file]
		} else {
			fileSize = fileInfo.Size()
		}

		File_Info = append(File_Info, FileInfo{
			Name: filepath.Base(file),
			Type: fileType,
			Size: fileSize,
		})
	}
	if sortOrder == "" {
		sortBySizeAsc(File_Info)
		fmt.Println("Способ сортировки не был выбран - выполняется сортировка по умолчанию (asc)")

	} else if sortOrder == "asc" {
		sortBySizeAsc(File_Info)
	} else if sortOrder == "desc" {
		sortBySizeDesc(File_Info)
	}
	for _, fileInfo := range File_Info {
		var sizeStr string
		switch {
		case fileInfo.Size >= GB:
			sizeStr = fmt.Sprintf("%.2f GB", float64(fileInfo.Size)/(thousand*thousand*thousand))
		case fileInfo.Size >= MB:
			sizeStr = fmt.Sprintf("%.2f MB", float64(fileInfo.Size)/(thousand*thousand))
		case fileInfo.Size >= KB:
			sizeStr = fmt.Sprintf("%.2f KB", float64(fileInfo.Size)/thousand)
		default:
			sizeStr = fmt.Sprintf("%d bytes", fileInfo.Size)
		}

		output := fmt.Sprintf("Name: %s, Type: %s, Size: %s", fileInfo.Name, fileInfo.Type, sizeStr)
		fmt.Println(output)
	}

	elapsed := time.Since(start)
	fmt.Printf("Время выполнения программы: %s\n", elapsed)
	return File_Info
}
