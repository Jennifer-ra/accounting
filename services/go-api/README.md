# Accounting Go API

Go/Gin backend for the Accounting Mini Program.

## Stack

- Go 1.23+
- Gin
- GORM
- MySQL 8.4 LTS
- WeChat Mini Program `jscode2session` login
- HMAC signed session token

## Quick Start

```bash
cd services/go-api
cp .env.example .env
docker compose up -d mysql
go mod tidy
go run ./cmd/server
```

Run the smoke test after the server starts:

```powershell
.\scripts\smoke.ps1
```

The API listens on `http://localhost:10240` by default.

For local development, `.env.example` enables `WECHAT_MOCK_LOGIN=true`, so the backend uses a stable mock openid. Disable this in production and configure `MINIPROGRAM_APPID` and `MINIPROGRAM_SECRET`.

## Environment

| Name | Required | Description |
| --- | --- | --- |
| `MYSQL_DSN` | yes | MySQL DSN. Use MySQL 8.4 LTS. |
| `SESSION_TOKEN_SECRET` | yes | HMAC signing secret for backend tokens. |
| `MINIPROGRAM_APPID` | production | WeChat Mini Program AppID. |
| `MINIPROGRAM_SECRET` | production | WeChat Mini Program Secret. |
| `WECHAT_MOCK_LOGIN` | local only | Use stable mock openid for local development. |
| `PORT` | no | API port, defaults to `10240`. |

## API

- `GET /health`
- `POST /api/v1/auth/wechat-login`
- `GET /api/v1/bootstrap?month=YYYY-MM`
- `GET /api/v1/transactions?month=YYYY-MM`
- `POST /api/v1/transactions`
- `PUT /api/v1/transactions/:id`
- `DELETE /api/v1/transactions/:id`
- `GET /api/v1/summary?month=YYYY-MM`
- `GET /api/v1/categories`
- `POST /api/v1/categories`
- `GET /api/v1/accounts`
- `POST /api/v1/accounts`
- `PUT /api/v1/budgets/:month`

## Database

The service uses GORM AutoMigrate at startup. A matching MySQL 8.4 reference schema is also kept in `migrations/001_init.sql` for review and manual deployment.

## Design Notes

The backend creates one default personal book per user at login. Every query is scoped by that book. This keeps the MVP small while preserving a clean extension path for shared/family books.

Transaction creation accepts a `clientId`. The pair `(book_id, client_id)` is unique, so a retried Mini Program request will not create duplicate transactions.

