# Order create acceptance: user :8081, inventory :8088, order :8089
$User = "http://127.0.0.1:8081"
$Inv = "http://127.0.0.1:8088"
$Ord = "http://127.0.0.1:8089"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

Write-Host "=== login ===" -ForegroundColor Cyan
$email = "ord-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
$regBody = "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"ord`"}"
Invoke-RestMethod -Uri "$User/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json; charset=utf-8" | Out-Null
$login = Invoke-RestMethod -Uri "$User/api/v1/auth/login" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`"}" -ContentType "application/json; charset=utf-8"
$token = $login.access_token
$headers = @{ Authorization = "Bearer $token" }
Assert "token" ($null -ne $token) "no token"

$sku = "01ORDERACCEPTSKU0000000001"
$initBody = "{`"sku_id`":`"$sku`",`"available`":10}"
try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories" -Method Post -Body $initBody -ContentType "application/json; charset=utf-8" | Out-Null
} catch { }
$before = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
$beforeAvail = [int]$before.data.available

Write-Host "`n=== create order ===" -ForegroundColor Cyan
$itemBody = "{`"items`":[{`"sku_id`":`"$sku`",`"product_id`":`"01ARZ3NDEKTSV4RRFFQ69G5FAW`",`"title`":`"Phone`",`"price_cent`":100,`"quantity`":2}]}"
$created = Invoke-RestMethod -Uri "$Ord/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers ($headers + @{ "Idempotency-Key" = "order-accept-1" })
Assert "status 10" ($created.data.status -eq 10) ($created | ConvertTo-Json -Compress)
Assert "pay 200" ($created.data.pay_amount_cent -eq 200) ($created.data.pay_amount_cent)
Assert "reservation" ($created.data.reservation_id.Length -eq 26) ($created.data.reservation_id)
$orderID = $created.data.id

$stock = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "available decreased" ($stock.data.available -eq ($beforeAvail - 2)) ($stock.data.available)
Assert "reserved 2" ($stock.data.reserved -eq 2) ($stock.data.reserved)

Write-Host "`n=== idempotency ===" -ForegroundColor Cyan
$again = Invoke-RestMethod -Uri "$Ord/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers ($headers + @{ "Idempotency-Key" = "order-accept-1" })
Assert "same order" ($again.data.id -eq $orderID) ($again.data.id)
$stock2 = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "stock unchanged" ($stock2.data.available -eq ($beforeAvail - 2)) ($stock2.data.available)

Write-Host "`n=== done ===" -ForegroundColor Cyan
if ($Fail -eq 0) { Write-Host "order create PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
