# Timeout close: unpaid order becomes status 40; stock released.
# Requires user :8081, inventory :8088, order :8089 with short TTL.
# Set MINIMALL_ORDER_PAYMENT_TTL=8 and MINIMALL_ORDER_CLOSE_INTERVAL_SEC=2 before starting order.exe
$ErrorActionPreference = "Stop"
$User = "http://127.0.0.1:8081"
$Inv = "http://127.0.0.1:8088"
$Ord = "http://127.0.0.1:8089"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

$email = "close-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
Invoke-RestMethod -Uri "$User/api/v1/auth/register" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"c`"}" -ContentType "application/json; charset=utf-8" | Out-Null
$login = Invoke-RestMethod -Uri "$User/api/v1/auth/login" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`"}" -ContentType "application/json; charset=utf-8"
$headers = @{ Authorization = "Bearer $($login.access_token)"; "Idempotency-Key" = "close-1" }

$sku = "01ORDERCLOSETESTSKU0000001" # CHAR(26)
try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories" -Method Post -Body "{`"sku_id`":`"$sku`",`"available`":30}" -ContentType "application/json; charset=utf-8" | Out-Null
} catch { }
$before = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
$beforeAvail = [int]$before.data.available
$beforeReserved = [int]$before.data.reserved

$itemBody = "{`"items`":[{`"sku_id`":`"$sku`",`"product_id`":`"01ARZ3NDEKTSV4RRFFQ69G5FAW`",`"title`":`"Close`",`"price_cent`":100,`"quantity`":2}]}"
$created = Invoke-RestMethod -Uri "$Ord/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers $headers
$orderID = [string]$created.data.id
Assert "pending" ($created.data.status -eq 10) $orderID
Assert "pay_expire_at set" ($created.data.pay_expire_at -ne $null -and "$($created.data.pay_expire_at)" -ne "") $created.data.pay_expire_at

$mid = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "reserved after create" ([int]$mid.data.reserved -eq ($beforeReserved + 2)) $mid.data.reserved

Write-Host "waiting for close worker (TTL ~8s)..."
$status = "10"
for ($i = 0; $i -lt 30; $i++) {
    Start-Sleep -Seconds 1
    $raw = cmd /c "docker compose -f deployments/docker-compose.yml exec -T mysql mysql -N -uminimall -p123456 -e `"SELECT status FROM minimall_order.orders WHERE id='$orderID'`" 2>nul"
    $status = (($raw | Out-String) -replace "[^\d]", "").Trim()
    if ($status -eq "40") { break }
}
Assert "status 40 closed" ($status -eq "40") $status

$after = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "available restored" ([int]$after.data.available -eq $beforeAvail) $after.data.available
Assert "reserved restored" ([int]$after.data.reserved -eq $beforeReserved) $after.data.reserved

if ($Fail -gt 0) {
    Write-Host "FAILED count: $Fail" -ForegroundColor Red
    exit 1
}
Write-Host "order timeout close PASSED" -ForegroundColor Green
