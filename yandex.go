package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func doYandexRequests() {
	var reportText strings.Builder

	dateStr := time.Now().Format("02.01.2006")

	reportText.WriteString(fmt.Sprintf("🟠 *Отчёт Yandex* [%s]:\n\n", dateStr))

	for accountName, feedID := range yandexFeeds {
		statusText := doYandexRequest(accountName, feedID)
		reportText.WriteString(statusText + "\n")
	}

	sendMessage(reportText.String())
}

func doYandexRequest(accountName, feedID string) string {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.realty.yandex.net/2.0/crm/feed/%s/state", feedID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Ошибка запроса Yandex для %s: %v", accountName, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", yandexOAuth))
	req.Header.Set("X-Authorization", "Vertis crm-dff153a8ef1a90d3bff5ee378dee416606cf8915")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Ошибка запроса Yandex для %s: %v", accountName, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Ошибка чтения ответа Yandex для %s: %v", accountName, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Ошибка парсинга JSON Yandex для %s: %v", accountName, err)
	}

	fileBase := strings.ToLower(strings.ReplaceAll(accountName, " ", "_"))

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("❌ *Yandex %s*: ошибка получения данных (код %d)", accountName, resp.StatusCode)
		writeJSONToFile(fmt.Sprintf("%s_yandex_response.json", fileBase), result)
		return msg
	}


	errorsList := []interface{}{}
	if state, ok := result["state"].(map[string]interface{}); ok {
		if errs, ok := state["errors"].([]interface{}); ok {
			errorsList = errs
		}
	}

	var msg string

	if len(errorsList) == 0 {
		msg = fmt.Sprintf("✅ *Yandex %s*: ошибок не найдено", accountName)
	} else {
		msg = fmt.Sprintf("🚫 *Yandex %s*: обнаружено %d ошибок", accountName, len(errorsList))

		errFile := fmt.Sprintf("%s_yandex_errors.json", fileBase)
		writeJSONToFile(errFile, errorsList)
		deleteFile(errFile)
	}

	writeJSONToFile(fmt.Sprintf("%s_yandex_response.json", fileBase), result)
	deleteFile(fmt.Sprintf("%s_yandex_response.json", fileBase))

	return msg
}
