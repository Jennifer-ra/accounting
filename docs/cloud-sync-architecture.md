# Cloud Sync Architecture

The project now supports two modes:

- Local mode: the Mini Program reads and writes WeChat local storage only.
- Cloud mode: the Mini Program logs in with `wx.login`, calls the Go API, and stores a local snapshot for fast rendering and fallback.

## Runtime Flow

```text
WeChat Mini Program
  wx.login()
      ↓
Go API /api/v1/auth/wechat-login
      ↓
WeChat jscode2session or local mock login
      ↓
users + default personal book
      ↓
HMAC session token
      ↓
Mini Program API calls with Authorization: Bearer <token>
```

## Data Boundaries

Every business table is scoped by `book_id`. The MVP creates one default personal book per user, while `book_members` keeps the path open for family/shared books later.

## Tables

- `users`: WeChat user identity. Stores `openid`; keeps `unionid` nullable for future multi-app identity linking.
- `books`: accounting books. The MVP uses one personal book.
- `book_members`: membership and roles for future shared books.
- `categories`: income/expense categories scoped by book.
- `accounts`: cash, WeChat Pay, Alipay, cards, and custom accounts scoped by book.
- `transactions`: income/expense records. Amounts are stored as integer cents.
- `budgets`: monthly budget per book.

## Query Performance

Important indexes:

- `users.open_id` unique index for login.
- `book_members(book_id, user_id)` unique index for future permission checks.
- `categories(book_id, type, name)` unique index for idempotent creation.
- `accounts(book_id, name)` unique index for idempotent creation.
- `transactions(book_id, client_id)` unique index for idempotent Mini Program retries.
- `transactions(book_id, tx_date)` index for monthly list and summary queries.
- `budgets(book_id, month)` unique index for monthly budget lookup.

## Idempotency

The Mini Program sends a `clientId` when creating transactions. The backend enforces `(book_id, client_id)` uniqueness and returns the existing transaction if the same request is retried.

## Local-To-Cloud Migration

When cloud mode is enabled, the Mini Program:

1. Logs in and gets a backend token.
2. Uploads local categories, accounts, and current month budget.
3. Uploads local bills that do not have a numeric `serverId`.
4. Pulls cloud transactions for the current month.
5. Stores the result back to local storage as a snapshot.

## Production Notes

- Disable `WECHAT_MOCK_LOGIN`.
- Set `MINIPROGRAM_APPID` and `MINIPROGRAM_SECRET`.
- Use a strong `SESSION_TOKEN_SECRET`.
- Put the API behind HTTPS.
- Add the API domain to WeChat Mini Program request legal domains.
- Use managed MySQL backups before inviting real users.
