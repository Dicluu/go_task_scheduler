## Notification
<h4> Notification - микросервис отвечающий за уведомление пользователей по почте в обозначеное время в заметке </h4>

## Конфигурация
Конфигурация предоставлена в /config/local.yaml.example

1) Скопировать local.yaml: `cp local.yaml.example local.yaml`

## Параметры
* env - режим работы приложения (local, dev, prod), влияет только на логирование
* storage_path - путь до файла бд sqlite
* address - адрес приложения в сети
* port - порт приложения в сети
* timeout - время для обработки реквеста и респонса
* idle_timeout - время жизни keepalive соединения
* username - логин почты
* password - временный пароль почты
* host - хост smtp
* port - порт smtp
* from - имя отправителя в сообщении
* is_tls - использование tls
* servers - адреса реплик kafka

## Применение миграций
1) запустить команду из корня микросервиса: `go run ./cmd/migrator/main.go --storage-path=[путь до файла бд] --migrations-path=./migrations/`

## Запуск
Выполнить последовательно следующие команды:
1) `export CONFIG_PATH=./config/local.yaml`
2) `go run ./cmd/main/main.go`