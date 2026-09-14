# Accounting

Accounting is a collection of personal accounting products and experiments. It starts with a WeChat Mini Program and now includes a Go API service for cloud sync.

## Projects

- `apps/wx-accounting-mvp`: WeChat Mini Program for daily expense and income tracking.
- `services/go-api`: Go/Gin backend with WeChat login, MySQL persistence, budgets, categories, accounts, and transaction APIs.

## Screenshots

![Home](docs/assets/wx-home.png)

## Features

- Fast expense and income entry from the home page
- Pure local mode for offline demos and personal experiments
- Cloud mode with WeChat login and backend persistence
- Local snapshot cache after cloud sync
- Monthly income, expense, balance, and budget progress
- Bill list with monthly filtering, editing, and deletion
- Category statistics for income and expense
- Custom categories, accounts, and monthly budget
- JSON export/import through the clipboard
- Heuristic CSV import for common payment bill formats

## Quick Start

Start MySQL 8.4 and the Go API:

```bash
cd services/go-api
cp .env.example .env
docker compose up -d mysql
go mod tidy
go run ./cmd/server
```

After the server starts, smoke-test the API:

```powershell
cd services/go-api
.\scripts\smoke.ps1
```

Import the Mini Program:

1. Open WeChat DevTools.
2. Import `apps/wx-accounting-mvp` as the project directory.
3. Use a test AppID or your own Mini Program AppID.
4. Compile and preview.
5. Open Settings, enable cloud mode, keep API URL as `http://localhost:10240/api/v1` for local development, and tap login/sync.

## Data Storage

Local mode uses WeChat local storage:

- `wx_accounting_mvp_bills`: bill records
- `wx_accounting_mvp_config`: categories, accounts, and monthly budget
- `wx_accounting_mvp_cloud_settings`: cloud mode and API settings
- `wx_accounting_mvp_token`: backend session token
- `wx_accounting_mvp_version`: last local update timestamp

Cloud mode stores canonical data in MySQL through the Go API, then caches the latest synced data locally for fast rendering.

## Architecture

See `docs/cloud-sync-architecture.md`.

## Roadmap

- Stable WeChat Pay and Alipay CSV import adapters
- Category and account sorting/deletion
- Recurring bills
- Shared/family books
- AI-assisted category detection and natural language bookkeeping

## License

MIT

