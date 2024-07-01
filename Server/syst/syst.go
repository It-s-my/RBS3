package syst

import (
	"fmt"
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

const (
	ASC  = "asc"
	DESC = "desc"
)

// FileInfo Структура, в которой будут записаны название, тип и размер файлов.
type FileInfo struct {
	Name  string
	Type  string
	Size  string
	Bsize int64
}

//sortBySizeAsc - принимает срез файлов для сортировки от меньшего к большему.
func sortBySizeAsc(files []FileInfo) []FileInfo {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Bsize < files[j].Bsize
	})
	return files
}

//sortBySizeDesc - принимает срез файлов и элементы сортируются в порядке убывания размера.
func sortBySizeDesc(files []FileInfo) []FileInfo {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Bsize > files[j].Bsize
	})
	return files
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

func Sortfile(rootPath string, sortOrder string) []FileInfo {

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
			Name:  filepath.Base(file),
			Type:  fileType,
			Bsize: fileSize,
		})
	}

	switch sortOrder {
	case "":
		File_Info = sortBySizeAsc(File_Info)
		fmt.Println("Способ сортировки не был выбран - выполняется сортировка по умолчанию (asc)")
		break
	case ASC:
		File_Info = sortBySizeAsc(File_Info)
		break
	case DESC:
		File_Info = sortBySizeDesc(File_Info)
		break
	}

	for i := 0; i < len(File_Info); i++ {
		var sizeStr string
		switch {
		case File_Info[i].Bsize >= GB:
			sizeStr = fmt.Sprintf("%.2f GB", float64(File_Info[i].Bsize)/(thousand*thousand*thousand))
		case File_Info[i].Bsize >= MB:
			sizeStr = fmt.Sprintf("%.2f MB", float64(File_Info[i].Bsize)/(thousand*thousand))
		case File_Info[i].Bsize >= KB:
			sizeStr = fmt.Sprintf("%.2f KB", float64(File_Info[i].Bsize)/thousand)
		default:
			sizeStr = fmt.Sprintf("%.2f  bytes", float64(File_Info[i].Bsize))
		}
		File_Info[i].Size = sizeStr
	}

	elapsed := time.Since(start)
	fmt.Printf("Время выполнения программы: %s\n", elapsed)
	return File_Info
}
