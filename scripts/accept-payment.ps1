# Pay flow: user :8081, inventory :8088, order :8089, payment :8085
$User = "http://127.0.0.1:8081"
$Inv = "http://127.0.0.1:8088"
$Ord = "http://127.0.0.1:8089"
$Pay = "http://127.0.0.1:8085"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

$email = "pay-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
$regBody = "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"p`"}"
Invoke-RestMethod -Uri "$User/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json; charset=utf-8" | Out-Null
$login = Invoke-RestMethod -Uri "$User/api/v1/auth/login" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`"}" -ContentType "application/json; charset=utf-8"
$headers = @{ Authorization = "Bearer $($login.access_token)" }

$sku = "01PAYFLOWTESTSKU0000000001"
try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories" -Method Post -Body "{`"sku_id`":`"$sku`",`"available`":20}" -ContentType "application/json; charset=utf-8" | Out-Null
} catch { }
$before = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
$beforeAvail = [int]$before.data.available

$h1 = $headers.Clone(); $h1["Idempotency-Key"] = "pay-order-1"
$itemBody = "{`"items`":[{`"sku_id`":`"$sku`",`"product_id`":`"01ARZ3NDEKTSV4RRFFQ69G5FAW`",`"title`":`"Phone`",`"price_cent`":100,`"quantity`":2}]}"
$created = Invoke-RestMethod -Uri "$Ord/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers $h1
$orderID = $created.data.id
Assert "order pending" ($created.data.status -eq 10) $orderID

$mid = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
$reservedMid = [int]$mid.data.reserved

$h2 = $headers.Clone(); $h2["Idempotency-Key"] = "pay-1"
$payBody = "{`"order_id`":`"$orderID`",`"amount_cent`":200}"
$payment = Invoke-RestMethod -Uri "$Pay/api/v1/payments" -Method Post -Body $payBody -ContentType "application/json; charset=utf-8" -Headers $h2
Assert "payment pending" ($payment.data.status -eq 10) ($payment | ConvertTo-Json -Compress)
$payID = $payment.data.id

$confirmed = Invoke-RestMethod -Uri "$Pay/api/v1/payments/$payID/mock-confirm" -Method Post -Headers $headers
Assert "payment paid" ($confirmed.data.status -eq 20) ($confirmed | ConvertTo-Json -Compress)

$after = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "available held" ([int]$after.data.available -eq ($beforeAvail - 2)) ($after.data.available)
Assert "reserved committed" ([int]$after.data.reserved -eq ($reservedMid - 2)) ($after.data.reserved)

Invoke-RestMethod -Uri "$Pay/api/v1/payments/$payID/mock-confirm" -Method Post -Headers $headers | Out-Null
$again = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "replay stable" (([int]$again.data.available -eq [int]$after.data.available) -and ([int]$again.data.reserved -eq [int]$after.data.reserved)) ($again.data | ConvertTo-Json -Compress)

$status = docker compose -f deployments/docker-compose.yml exec -T mysql mysql -N -uminimall -p123456 -e "SELECT status FROM minimall_order.orders WHERE id='$orderID'"
$status = ($status | Out-String).Trim()
Assert "order status 20" ($status -eq "20") $status

if ($Fail -eq 0) { Write-Host "payment flow PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
