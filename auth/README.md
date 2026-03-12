## Auth
<h4> Auth - микросервис отвечающий за регистрацию, авторизацию и рефреш JWT токенов </h4>

## Конфигурация
Конфигурация предоставлена в /config/local.yaml.example

1) Скопировать local.yaml: `cp local.yaml.example local.yaml`

## Параметры
* env - режим работы приложения (local, dev, prod), влияет только на логирование
* storage_path - путь до файла бд sqlite
* token_ttl - время жизни jwt access token
* refresh_token_ttl - время жизни jwt refresh token
* secret - секретный ключ для подписи jwt
* address - адрес приложения в сети
* timeout - время для обработки реквеста и респонса
* idle_timeout - время жизни keepalive соединения
* servers - адреса реплик kafka

## Применение миграций
1) запустить команду из корня микросервиса: `go run ./cmd/migrator/main.go --storage-path=[путь до файла бд] --migrations-path=./migrations/`

## Запуск
Выполнить последовательно следующие команды:
1) `export CONFIG_PATH=./config/local.yaml`
2) `go run ./cmd/main/main.go`