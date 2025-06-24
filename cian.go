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

func doCianRequests() {
	var reportText strings.Builder

	dateStr := time.Now().Format("02.01.2006")

	reportText.WriteString(fmt.Sprintf("🔵 *Отчёт Cian* [%s]:\n\n", dateStr))

	for accountName, token := range accounts {
		statusText := doCianRequest(accountName, token)
		reportText.WriteString(statusText + "\n")
	}

	sendMessage(reportText.String())
}

func doCianRequest(accountName, token string) string {
	client := &http.Client{}

	req, err := http.NewRequest("GET", cianURL, nil)
	if err != nil {
		log.Printf("Ошибка создания запроса Cian: %v", err)
		return fmt.Sprintf("❌ *Cian %s*: ошибка создания запроса", accountName)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Ошибка выполнения запроса Cian для %s: %v", accountName, err)
		return fmt.Sprintf("❌ *Cian %s*: ошибка выполнения запроса", accountName)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Ошибка чтения ответа Cian для %s: %v", accountName, err)
		return fmt.Sprintf("❌ *Cian %s*: ошибка чтения ответа", accountName)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Printf("Ошибка парсинга JSON Cian для %s: %v", accountName, err)
		return fmt.Sprintf("❌ *Cian %s*: ошибка парсинга JSON", accountName)
	}

	fileBase := strings.ToLower(strings.ReplaceAll(accountName, " ", "_"))

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("❌ *Cian %s*: ошибка получения данных (код %d)", accountName, resp.StatusCode)
		writeJSONToFile(fmt.Sprintf("%s_response.json", fileBase), data)
		deleteFile(fmt.Sprintf("%s_response.json", fileBase))
		return msg
	}

	offers := extractOffers(data)

	errorsList := filterOffersByKey(offers, "errors")
	warningsList := filterOffersByKey(offers, "warnings")

	var msg string

	if len(errorsList) == 0 && len(warningsList) == 0 {
		msg = fmt.Sprintf("✅ *Cian %s*: ошибок и предупреждений не найдено", accountName)
	} else {
		if len(errorsList) > 0 {
			msg += fmt.Sprintf("🚫 %d ошибок", len(errorsList))

			errFile := fmt.Sprintf("%s_errors.json", fileBase)
			writeJSONToFile(errFile, errorsList)
			deleteFile(errFile)
		}
		if len(warningsList) > 0 {
			if len(msg) > 0 {
				msg += ", "
			}
			msg += fmt.Sprintf("⚠️ %d предупреждений", len(warningsList))

			warnFile := fmt.Sprintf("%s_warnings.json", fileBase)
			writeJSONToFile(warnFile, warningsList)
			deleteFile(warnFile)
		}

		msg = fmt.Sprintf("🚫 *Cian %s*: %s", accountName, msg)
	}

	deleteFile(fmt.Sprintf("%s_response.json", fileBase))

	return msg
}