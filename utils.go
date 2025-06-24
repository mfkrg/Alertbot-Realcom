package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func extractOffers(data map[string]interface{}) []map[string]interface{} {
	offers := []map[string]interface{}{}
	if result, ok := data["result"].(map[string]interface{}); ok {
		if offersList, ok := result["offers"].([]interface{}); ok {
			for _, item := range offersList {
				if offer, ok := item.(map[string]interface{}); ok {
					offers = append(offers, offer)
				}
			}
		}
	}
	return offers
}

func filterOffersByKey(offers []map[string]interface{}, key string) []map[string]interface{} {
	filtered := []map[string]interface{}{}
	for _, offer := range offers {
		if v, ok := offer[key]; ok {
			if arr, ok := v.([]interface{}); ok && len(arr) > 0 {
				filtered = append(filtered, offer)
			}
		}
	}
	return filtered
}

func writeJSONToFile(filename string, data interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Ошибка создания файла %s: %v", filename, err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		log.Printf("Ошибка записи JSON в %s: %v", filename, err)
	}
}

func deleteFile(filePath string) {
	err := os.Remove(filePath)
	if err != nil {
		log.Printf("Ошибка удаления файла %s: %v", filePath, err)
	} else {
		log.Printf("Удален файл: %s", filePath)
	}
}

func cleanupJSONFiles() {
	files, err := filepath.Glob("*.json")
	if err != nil {
		log.Printf("Ошибка поиска JSON файлов: %v", err)
		return
	}

	for _, file := range files {
		if strings.HasPrefix(file, "avito_token_") {
			continue
		}

		err := os.Remove(file)
		if err != nil {
			log.Printf("Ошибка удаления файла %s: %v", file, err)
		} else {
			log.Printf("Удален файл: %s", file)
		}
	}
}