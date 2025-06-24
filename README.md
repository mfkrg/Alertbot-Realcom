# 🚀 AlertBot

Приложение на Go для автоматизированной проверки выгрузки через API площадок (Циан, Авито, Яндекс) с запуском по расписанию внутри Docker-контейнера через встроенный cron.

---

## 📦 Состав проекта

- `main.go` — entry-point приложения
- `Dockerfile` — файл для развертывания в Docker
- `docker-compose.yml` — Compose-файл проекта
- `entrypoint.sh` — запуск cron внутри контейнера
- `cronfile` — расписание работы
- `yandex.go` — обработка площадки Яндекс
- `cian.go` — обработка площадки Циан
- `avito.go` — обработка площадки Авито
- `utils.go` — функции работы с файлами
- `telegram.go` — обработчик для Telegram-бота

---

## ⚙️ Сборка контейнера

```bash
docker-compose up --build
```

---
## 📝 Первый запуск
### 1. Создание файла config.go
```bash
Пример заполнения файла
package main

// YANDEX

var yandexFeeds = map[string]string{
	"КАБИНЕТ 1":              "FEED ID",
	"КАБИНЕТ 2":              "FEED ID",
}

const yandexOAuth = "Ключ авторизации Яндекс"

// CIAN

var accounts = map[string]string{
	"КАБИНЕТ 1":          "API KEY CIAN",
	"КАБИНЕТ 2":          "API KEY CIAN",
}

const cianURL = "Обрабатываемый URL API"


// AVITO

var avitoAccounts = map[string]AvitoAccount{
	"КАБИНЕТ !": {
		ClientID:     "CLIENT ID",
		ClientSecret: "CLIENT SECRET",
	},
	"КАБИНЕТ 2": {
		ClientID:     "CLIENT ID",
		ClientSecret: "CLIENT SECRET",
	},
}

const botToken = "TELEGRAM BOT TOKEN"
const channelID = "CHAT ID"
```

### 2. Создание cronfile
```bash
0 10 * * * docker run --rm alertbot >> /var/log/alertbot.log 2>&1

```

### 3. Сборка проекта (git bash)
```bash
GOOS=операционная_система GOARCH=архитектура_процессора go build -o название_файла .
```
---
# FAQ
**Q: Что вызывает cronfile?**  
A: /app/alertbot.

**Q: Как изменить расписание?**  
A: Скорректировать cron-файл

**Q: Где сохраняются логи?**  
A: Логи крон - /var/log/cron.log  
Логи приложения - /var/log/alertbot.log