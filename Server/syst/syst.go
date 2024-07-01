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

// sortBySizeAsc - принимает срез файлов для сортировки от меньшего к большему.
func sortBySizeAsc(files []FileInfo) []FileInfo {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Bsize < files[j].Bsize
	})
	return files
}

// sortBySizeDesc - принимает срез файлов и элементы сортируются в порядке убывания размера.
func sortBySizeDesc(files []FileInfo) []FileInfo {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Bsize > files[j].Bsize
	})
	return files
}

// walk - принимает строку root которая содержит информацию о размере каждой директории, начиная с корневой директории root.
func walk(root string) (map[string]int64, error) {

	// directorySizes пустая карта, которая будет содержать путь к директории и ее размер.
	directorySizes := make(map[string]int64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	//Функция используется для рекурсивного прохода по директориям, начиная с корневой директории root.
	//Принимает путь к директории, функцию обратного вызова и возвращает ошибку, если что-то пошло не так.
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error { //
		if err != nil {
			fmt.Printf("Ошибка доступа %q: %v\n", path, err)
			return err
		}
		//Условие, если  на пути папка, то запускается горутина, которая обрабатывает эту папку
		//Для безопасности доступа к общей переменной directorySizes используется Lock и Unlock.
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

	// walk - возвращает карту directorySizes, содержащую информацию о размере каждой директории и ошибку.
	return directorySizes, nil
}

// GetFailesList принимает путь и тип сортировки, возвращая FileInfo, в котором содержится вся информация о директории.
func GetFailesList(rootPath string, sortOrder string) ([]FileInfo, error) {
	//Запуск таймера выполнения программы
	start := time.Now()

	//Происходит получение списка файлов в указанной директории, используя filepath.Glob.
	//Эта функция возвращает список файлов в указанной директории, соответствующих шаблону (все файлы).
	files, err := filepath.Glob(filepath.Join(rootPath, "*"))
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	FilesInfo := make([]FileInfo, 0)

	//Вызывается функция walk, которая рекурсивно обходит все поддиректории начиная с указанной корневой директории.
	//Функция возвращает карту, где ключами являются пути к директориям, а значениями - их размеры. А так же возвращается ошибка и обрабатывается.
	directories, walkErr := walk(rootPath)
	if walkErr != nil {
		fmt.Println("Ошибка при обходе файловой системы:", walkErr)
		return nil, walkErr
	}

	//Происходит чтение информации о файлах
	//Для каждого файла вызывается os.Stat, чтобы получить информацию о файле (размер, тип и имя).
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			fmt.Println("Ошибка чтения информации о файле:", err)
			continue
		}
		//Определяется тип файла (файл или директория) и размер файла.
		fileType := "file"
		if fileInfo.IsDir() {
			fileType = "directory"
		}
		// Получение размера папки, если это директория
		var fileSize int64
		if fileInfo.IsDir() {
			fileSize = directories[file]
		} else {
			fileSize = fileInfo.Size()
		}
		// File_Info содержит информацию о всех файлах и директориях в указанной директории имя,тип и размеры.
		FilesInfo = append(FilesInfo, FileInfo{
			Name:  filepath.Base(file),
			Type:  fileType,
			Bsize: fileSize,
		})
	}

	//Сортировка файлов
	switch sortOrder {
	case "":
		FilesInfo = sortBySizeAsc(FilesInfo)
		fmt.Println("Способ сортировки не был выбран - выполняется сортировка по умолчанию (asc)")
		break
	case ASC:
		FilesInfo = sortBySizeAsc(FilesInfo)
		break
	case DESC:
		FilesInfo = sortBySizeDesc(FilesInfo)
		break
	}

	//Конвертация размера из байтов в нормальные значения
	for i := 0; i < len(FilesInfo); i++ {
		var sizeStr string
		switch {
		case FilesInfo[i].Bsize >= GB:
			sizeStr = fmt.Sprintf("%.2f GB", float64(FilesInfo[i].Bsize)/(thousand*thousand*thousand))
		case FilesInfo[i].Bsize >= MB:
			sizeStr = fmt.Sprintf("%.2f MB", float64(FilesInfo[i].Bsize)/(thousand*thousand))
		case FilesInfo[i].Bsize >= KB:
			sizeStr = fmt.Sprintf("%.2f KB", float64(FilesInfo[i].Bsize)/thousand)
		default:
			sizeStr = fmt.Sprintf("%.2f  bytes", float64(FilesInfo[i].Bsize))
		}
		FilesInfo[i].Size = sizeStr
	}
	//Остановка таймера
	elapsed := time.Since(start)
	fmt.Printf("Время выполнения программы: %s\n", elapsed)
	return FilesInfo, nil
}
