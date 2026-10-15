# M1 gateway acceptance: user :8081, product :8082, gateway :8080 all running
$Gw = "http://127.0.0.1:8080"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

Write-Host "=== healthz ===" -ForegroundColor Cyan
try {
    $hz = Invoke-RestMethod -Uri "$Gw/healthz" -Method Get
    Assert "healthz" ($hz.status -eq "ok") ($hz | ConvertTo-Json -Compress)
} catch { Assert "healthz" $false $_.Exception.Message; exit 1 }

Write-Host "`n=== register + login via gateway ===" -ForegroundColor Cyan
$email = "gw-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
$regBody = "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"gw`"}"
try {
    $reg = Invoke-RestMethod -Uri "$Gw/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json; charset=utf-8"
    Assert "register access_token" ($null -ne $reg.access_token) ($reg | ConvertTo-Json -Compress)
} catch {
    Assert "register" $false $_.Exception.Message
}

$loginBody = "{`"email`":`"$email`",`"password`":`"Pass123456`"}"
$login = Invoke-RestMethod -Uri "$Gw/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json; charset=utf-8"
Assert "login access_token" ($null -ne $login.access_token) ($login | ConvertTo-Json -Compress)
$token = $login.access_token
Assert "access_token" ($null -ne $token -and $token.Length -gt 10) "missing token"

$headers = @{ Authorization = "Bearer $token" }

Write-Host "`n=== /users/me ===" -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "$Gw/api/v1/users/me" -Method Get | Out-Null
    Assert "GET /users/me no token" $false "expected 401"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert "GET /users/me no token -> 401" ($code -eq 401) "status=$code"
}

$me = Invoke-RestMethod -Uri "$Gw/api/v1/users/me" -Method Get -Headers $headers
Assert "GET /users/me with token" ($null -ne $me.id) ($me | ConvertTo-Json -Compress)

Write-Host "`n=== public product read ===" -ForegroundColor Cyan
$prods = Invoke-RestMethod -Uri "$Gw/api/v1/products" -Method Get
Assert "GET /products no token" ($prods.code -eq 0) ($prods | ConvertTo-Json -Compress)

Write-Host "`n=== protected product write ===" -ForegroundColor Cyan
$prodBody = "{`"title`":`"GW Test`",`"status`":1}"
try {
    Invoke-RestMethod -Uri "$Gw/api/v1/products" -Method Post -Body $prodBody -ContentType "application/json; charset=utf-8" | Out-Null
    Assert "POST /products no token" $false "expected 401"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert "POST /products no token -> 401" ($code -eq 401) "status=$code"
}

$created = Invoke-RestMethod -Uri "$Gw/api/v1/products" -Method Post -Body $prodBody -ContentType "application/json; charset=utf-8" -Headers $headers
Assert "POST /products with token" ($created.code -eq 0) ($created | ConvertTo-Json -Compress)

Write-Host "`n=== done ===" -ForegroundColor Cyan
if ($Fail -eq 0) { Write-Host "M1 gateway PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
