
# Reportify

**Reportify** — web-система для создания аналитических табличных и графических отчётов **без написания SQL-запросов** со стороны пользователя.

Проект разработан как **pet-project / academic project** и предназначен для демонстрации архитектуры, работы с БД, безопасности и веб-разработки.

---

## Кратко о проекте

- Backend: Go
- Frontend: Vanilla JS + HTML/CSS
- DB: PostgreSQL, Redis
- Infra: Docker, Docker Compose, Nginx
- Архитектура: layered architecture (web → service → repository)
- Статус: завершён, не в промышленной эксплуатации

---

## Что делает система

- Позволяет создавать аналитические отчёты без ручного SQL
- Формирует табличные и графические отчёты (bar / line / pie)
- Управляет пользователями, ролями и правами доступа
- Хранит и переиспользует отчёты
- Предоставляет web-интерфейс и REST API

---

## Ключевые технические решения

### Backend

- Чёткое разделение слоёв:
  - `web` — HTTP, middleware, handlers
  - `service` — бизнес-логика
  - `repository` — доступ к данным
- Dependency Injection через явную инициализацию
- DTO для изоляции слоёв
- Тесты для сервисного слоя

### Безопасность

- Хэширование паролей через **BCrypt**
- Token-based аутентификация
- Конфигурируемое время жизни токенов
- Middleware для авторизации и логирования
- Автоматическая ротация секретов при перезапуске

### Работа с БД

- PostgreSQL
- Версионируемые SQL-миграции
- Разделение схем (identity / report / metadata)
- Репозиторный слой без утечек SQL в сервисы

### Frontend

- Vanilla JS без фреймворков
- Модульная структура
- Кастомный SQL-builder на стороне клиента
- Графики и UI-компоненты реализованы вручную

### Инфраструктура

- Полная контейнеризация
- Nginx как reverse proxy + static
- Makefile для локального запуска
- `.env`-конфигурация с валидацией

---

## Структура проекта (сокращённо)

```text
cmd/            — точка входа
internal/
  web/          — HTTP сервер, middleware
  service/      — бизнес-логика
  repository/   — работа с БД
  secure/       — хеширование, токены
  config/       — конфигурация
docker/         — compose, nginx, миграции
```



## Авторы

Backend Go developer - Евгений Егоров (https://github.com/ummuys)

Frontend developer - Никита Сорокин (https://github.com/nikitasoro-kin)

Frontend developer - Никита Сабиров (https://github.com/Ares-13)
