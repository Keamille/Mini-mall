# Cart acceptance via gateway :8080 (needs user + cart; product optional for real sku)
$Gw = "http://127.0.0.1:8080"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

Write-Host "=== login ===" -ForegroundColor Cyan
$email = "cart-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
$regBody = "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"cart`"}"
Invoke-RestMethod -Uri "$Gw/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json; charset=utf-8" | Out-Null
$login = Invoke-RestMethod -Uri "$Gw/api/v1/auth/login" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`"}" -ContentType "application/json; charset=utf-8"
$token = $login.access_token
$headers = @{ Authorization = "Bearer $token" }
Assert "access_token" ($null -ne $token) "no token"

$skuId = "01ARZ3NDEKTSV4RRFFQ69G5FAV"

Write-Host "`n=== GET cart no auth ===" -ForegroundColor Cyan
try {
    Invoke-RestMethod -Uri "$Gw/api/v1/cart" -Method Get | Out-Null
    Assert "GET cart 401" $false "expected 401"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert "GET cart 401" ($code -eq 401) "status=$code"
}

Write-Host "`n=== POST add item ===" -ForegroundColor Cyan
$addBody = "{`"sku_id`":`"$skuId`",`"quantity`":2}"
$add = Invoke-RestMethod -Uri "$Gw/api/v1/cart" -Method Post -Body $addBody -ContentType "application/json; charset=utf-8" -Headers $headers
Assert "POST cart" ($add.code -eq 0) ($add | ConvertTo-Json -Compress)

$list = Invoke-RestMethod -Uri "$Gw/api/v1/cart" -Method Get -Headers $headers
Assert "GET cart has item" ($list.data.Count -ge 1) ($list | ConvertTo-Json -Compress)
$row = $list.data | Where-Object { $_.sku_id -eq $skuId } | Select-Object -First 1
Assert "quantity=2" ($row.quantity -eq 2) ($row.quantity)

Write-Host "`n=== POST update qty ===" -ForegroundColor Cyan
$updBody = "{`"sku_id`":`"$skuId`",`"quantity`":5}"
Invoke-RestMethod -Uri "$Gw/api/v1/cart" -Method Post -Body $updBody -ContentType "application/json; charset=utf-8" -Headers $headers | Out-Null
$list2 = Invoke-RestMethod -Uri "$Gw/api/v1/cart" -Method Get -Headers $headers
$row2 = $list2.data | Where-Object { $_.sku_id -eq $skuId } | Select-Object -First 1
Assert "quantity=5" ($row2.quantity -eq 5) ($row2.quantity)

Write-Host "`n=== DELETE item ===" -ForegroundColor Cyan
Invoke-RestMethod -Uri "$Gw/api/v1/cart/$skuId" -Method Delete -Headers $headers | Out-Null
$list3 = Invoke-RestMethod -Uri "$Gw/api/v1/cart" -Method Get -Headers $headers
$left = @($list3.data | Where-Object { $_.sku_id -eq $skuId })
Assert "deleted" ($left.Count -eq 0) ($list3 | ConvertTo-Json -Compress)

Write-Host "`n=== done ===" -ForegroundColor Cyan
if ($Fail -eq 0) { Write-Host "cart PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
