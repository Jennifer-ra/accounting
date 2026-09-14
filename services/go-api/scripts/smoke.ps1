$ErrorActionPreference = 'Stop'
$baseUrl = $env:ACCOUNTING_API_BASE_URL
if (-not $baseUrl) { $baseUrl = 'http://localhost:10240' }

$login = Invoke-RestMethod -Method Post -Uri "$baseUrl/api/v1/auth/wechat-login" -ContentType 'application/json' -Body (@{ code = 'dev-code' } | ConvertTo-Json)
$token = $login.data.token
$headers = @{ Authorization = "Bearer $token" }

Invoke-RestMethod -Method Get -Uri "$baseUrl/health"
Invoke-RestMethod -Method Get -Uri "$baseUrl/api/v1/bootstrap?month=2026-09" -Headers $headers
Invoke-RestMethod -Method Put -Uri "$baseUrl/api/v1/budgets/2026-09" -Headers $headers -ContentType 'application/json' -Body (@{ amountFen = 300000 } | ConvertTo-Json)
Invoke-RestMethod -Method Post -Uri "$baseUrl/api/v1/transactions" -Headers $headers -ContentType 'application/json' -Body (@{
  clientId = 'smoke-2026-09-14-001'
  type = 'expense'
  amountFen = 2800
  category = '餐饮'
  account = '微信'
  date = '2026-09-14'
  note = 'smoke test'
  source = 'manual'
} | ConvertTo-Json)
Invoke-RestMethod -Method Get -Uri "$baseUrl/api/v1/summary?month=2026-09" -Headers $headers
