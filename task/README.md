## Task
<h4> Task - микросервис отвечающий за crud заметок и отправку уведомлений в назначенное время в заметке </h4>

## Конфигурация
Конфигурация предоставлена в /config/local.yaml.example

1) Скопировать local.yaml: `cp local.yaml.example local.yaml`

## Параметры
* env - режим работы приложения (local, dev, prod), влияет только на логирование
* storage_path - путь до файла бд sqlite
* secret - секретный ключ для подписи jwt
* address - адрес приложения в сети
* timeout - время для обработки реквеста и респонса
* idle_timeout - время жизни keepalive соединения
* address (cron) - адрес микросервиса нотификаций
* batch - количество записей в gRPC стриме для сервиса нотификаций

## Применение миграций
1) запустить команду из корня микросервиса: `go run ./cmd/migrator/main.go --storage-path=[путь до файла бд] --migrations-path=./migrations/`

## Запуск
Выполнить последовательно следующие команды:
1) `export CONFIG_PATH=./config/local.yaml`
2) `go run ./cmd/main/main.go`