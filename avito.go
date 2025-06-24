package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type AvitoAccount struct {
	ClientID     string
	ClientSecret string
}

type AvitoToken struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func doAvitoRequests() {
	var reportText strings.Builder

	dateStr := time.Now().Format("02.01.2006")

	reportText.WriteString(fmt.Sprintf("🟢 *Отчёт Avito* [%s]:\n\n", dateStr))

	for accountName, creds := range avitoAccounts {
		statusText := doAvitoRequestForAccount(accountName, creds.ClientID, creds.ClientSecret)
		reportText.WriteString(statusText + "\n")
	}

	sendMessage(reportText.String())
}

func doAvitoRequestForAccount(accountName, clientID, clientSecret string) string {
	log.Printf("=== Avito [%s] ===", accountName)

	token := loadAvitoTokenFromFile(accountName)

	if !isAvitoTokenValid(token) {
		log.Printf("[%s] Недействительный токен Avito — получение нового токена", accountName)
		var err error
		token, err = getAvitoToken(clientID, clientSecret)
		if err != nil {
			log.Printf("[%s] Ошибка получения токена Avito: %v", accountName, err)
			return fmt.Sprintf("❌ *Avito %s*: ошибка получения токена", accountName)
		}
		saveAvitoTokenToFile(accountName, token)
	}

	endpoint := "https://api.avito.ru/autoload/v3/reports/last_completed_report"

	tryRequest := func(token AvitoToken) (*http.Response, error) {
		client := &http.Client{}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}

	resp, err := tryRequest(token)
	if err != nil {
		log.Printf("[%s] Ошибка выполнения запроса Avito: %v", accountName, err)
		return fmt.Sprintf("❌ *Avito %s*: ошибка выполнения запроса", accountName)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		log.Printf("[%s] Получен 403 — обновление токена Avito...", accountName)
		token, err = getAvitoToken(clientID, clientSecret)
		if err != nil {
			log.Printf("[%s] Ошибка обновления токена Avito: %v", accountName, err)
			return fmt.Sprintf("❌ *Avito %s*: ошибка обновления токена", accountName)
		}
		saveAvitoTokenToFile(accountName, token)

		resp.Body.Close()

		resp, err = tryRequest(token)
		if err != nil {
			log.Printf("[%s] Ошибка после обновления токена: %v", accountName, err)
			return fmt.Sprintf("❌ *Avito %s*: ошибка после обновления токена", accountName)
		}
		defer resp.Body.Close()
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] Ошибка чтения ответа Avito: %v", accountName, err)
		return fmt.Sprintf("❌ *Avito %s*: ошибка чтения ответа", accountName)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("[%s] Ошибка парсинга JSON Avito: %v", accountName, err)
		return fmt.Sprintf("❌ *Avito %s*: ошибка парсинга JSON", accountName)
	}

	fileBase := strings.ToLower(strings.ReplaceAll(accountName, " ", "_"))

	writeJSONToFile(fmt.Sprintf("%s_avito_response.json", fileBase), result)

	errorCount := 0

	if sectionStats, ok := result["section_stats"].(map[string]interface{}); ok {
		if sections, ok := sectionStats["sections"].([]interface{}); ok {
			for _, sec := range sections {
				if section, ok := sec.(map[string]interface{}); ok {
					if slug, ok := section["slug"].(string); ok && slug == "error" {
						if count, ok := section["count"].(float64); ok {
							errorCount = int(count)
						}
					}
				}
			}
		}
	}

	var msg string
	if errorCount == 0 {
		msg = fmt.Sprintf("✅ *Avito %s*: ошибок не найдено", accountName)
	} else {
		msg = fmt.Sprintf("🚫 *Avito %s*: обнаружено %d ошибок", accountName, errorCount)
	}

	deleteFile(fmt.Sprintf("%s_avito_response.json", fileBase))

	return msg
}

func getAvitoToken(clientID, clientSecret string) (AvitoToken, error) {
	client := &http.Client{}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	req, err := http.NewRequest("POST", "https://api.avito.ru/token/", strings.NewReader(data.Encode()))
	if err != nil {
		return AvitoToken{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return AvitoToken{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AvitoToken{}, err
	}

	if resp.StatusCode != 200 {
		return AvitoToken{}, fmt.Errorf("Ошибка получения токена: status %d, body %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return AvitoToken{}, err
	}

	tokenStr, ok := result["access_token"].(string)
	if !ok {
		return AvitoToken{}, fmt.Errorf("access_token не найден в ответе")
	}

	token := AvitoToken{
		AccessToken: tokenStr,
		ExpiresAt:   time.Now().Add(23 * time.Hour),
	}

	log.Printf("Получен новый токен Avito")

	return token, nil
}

func isAvitoTokenValid(token AvitoToken) bool {
	return token.AccessToken != "" && time.Now().Before(token.ExpiresAt)
}

func loadAvitoTokenFromFile(accountName string) AvitoToken {
	fileName := fmt.Sprintf("avito_token_%s.json", strings.ToLower(strings.ReplaceAll(accountName, " ", "_")))

	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("[%s] Нет существующего файла токена — будет получен новый токен", accountName)
		return AvitoToken{}
	}
	defer file.Close()

	var token AvitoToken
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&token); err != nil {
		log.Printf("[%s] Ошибка декодирования файла токена: %v", accountName, err)
		return AvitoToken{}
	}

	log.Printf("[%s] Токен загружен из файла", accountName)
	return token
}

func saveAvitoTokenToFile(accountName string, token AvitoToken) {
	fileName := fmt.Sprintf("avito_token_%s.json", strings.ToLower(strings.ReplaceAll(accountName, " ", "_")))

	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("[%s] Ошибка сохранения токена: %v", accountName, err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&token); err != nil {
		log.Printf("[%s] Ошибка записи токена в файл: %v", accountName, err)
		return
	}

	log.Printf("[%s] Токен сохранен в файл", accountName)
}

