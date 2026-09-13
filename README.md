# Accounting

Personal accounting products and experiments. This repository starts with a local-first WeChat Mini Program MVP and is structured as a monorepo so backend services, import tools, documentation, and future clients can be added later.

## Projects

- `apps/wx-accounting-mvp`: local-first WeChat Mini Program for daily expense and income tracking.

## Current Data Storage

The Mini Program stores data in WeChat local storage on the current device/runtime:

- `wx_accounting_mvp_bills`: bill records
- `wx_accounting_mvp_config`: categories, accounts, and monthly budget
- `wx_accounting_mvp_version`: last local update timestamp

This means data is not synced across devices yet. Clearing Mini Program storage will remove local data unless it has been exported.

## Roadmap

- Stable WeChat Pay and Alipay CSV import adapters
- Category and account sorting/deletion
- Recurring bills
- Go API service for account sync
- AI-assisted category detection and natural language bookkeeping
