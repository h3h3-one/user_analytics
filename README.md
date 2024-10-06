# Сервис сбора и храниения действий пользователей

Стек: Go, Fiber, Logrus, PostgreSQL, Docker.

## Возможности конфигурирования

В файле docker-compose.yaml можно настроить под себя следующие переменные: 
1) LOG_LEVEL - выбор уровня логирования - FATAL, INFO, ERROR, DEBUG(по умолчанию)
2) POOL_COUNT - количество воркеров
3) DB_USER, DB_PASSWORD, DB_NAME - переменные для подключения к базе из контейнера с приложением
4) POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB - переменные для создания пользователя и базы из контейнера с PostrgeSQL

## Запуск приложения

1) Установить Docker и Golang version >=1.22
3) Перейти в папку с кодом и ввести команду:

```bash
go mod tidy
docker compose up --build
```


## Как пользоваться

Запускаем команду:

```bash
curl -location -request POST 'http://localhost:8080/analytics' --header 'X-Tantum-UserAgent: DeviceID=G1752G75-7C56-4G49-BGFA5ACBGC963471;DeviceType=iOS;OsVersion=15.5;AppVersion=4.3 (725)' --header 'X-Tantum-Authorization: 2daba111-1e48-4ba1-8753-2daba1119a09' --header 'Content-Type: application/json' --data-raw '{ "module" : "settings", "type" : "alert", "event" : "click", "name" : "подтверждение выхода", "data" : {"action" : "cancel"} }'
```

Должны получить ответ:

code 202
```json
{
	"status":"ok"
}
```

или

code 400
```json
{
	"message":"missing required headers"
}
```

Также можно открыть метрики приложения по муршруту `localhost:8080/metrics`