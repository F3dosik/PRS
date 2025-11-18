# Флаги
APP_NAME=prs
APP_PORT=8080

.PHONY: build up down logs lint migrate 

# Собираем бинарь внутри контейнера builder
build:
	docker-compose build app

# Поднимаем сервис и БД (миграции выполнены автоматически через сервис migrate)
up:
	docker-compose up -d

# Останавливаем сервисы
down:
	docker-compose down

# Очистить базу и тома
clean:
	docker-compose down -v

# Логи всех сервисов
logs:
	docker-compose logs -f

# Линтер 
lint:
	docker-compose run --rm lint

# Применение миграций вручную (если нужно)
migrate:
	docker-compose run --rm migrate

