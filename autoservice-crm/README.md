# AutoService CRM

MVP CRM-система для автосервиса: Go + PostgreSQL + HTML/CSS + Docker + HTTPS через Caddy.

## Быстрый запуск

cp .env.example .env
docker compose up --build

Открыть:

https://localhost

Браузер может показать предупреждение о локальном сертификате. Это нормально: Caddy использует внутренний TLS-сертификат.

## Возможности

- регистрация и авторизация;
- dashboard;
- учет клиентов;ы
- учет автомобилей;
- заказ-наряды;
- смена статуса заказ-наряда;
- PDF заказ-наряда;
- Docker Compose;
- HTTPS через Caddy.

## Структура

cmd/server/main.go       — backend
web/templates/           — HTML-шаблоны
web/static/css/app.css   — Apple-style CSS
migrations/001_init.sql  — структура БД
Dockerfile               — сборка Go-приложения
docker-compose.yml       — app + postgres + caddy
