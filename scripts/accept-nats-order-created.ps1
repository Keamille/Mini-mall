# Verify order_created reaches inbox via NATS backlog (gRPC notify will fail while notification is down).
# Requires: mysql, nats, user, inventory, order. Starts/stops notification.exe in repo root.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

$User = "http://127.0.0.1:8081"
$Inv = "http://127.0.0.1:8088"
$Ord = "http://127.0.0.1:8089"
$Notif = "http://127.0.0.1:8087"

# Stop notification so only NATS can deliver after restart.
Get-Process -Name "notification" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 1

$email = "nats-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
$reg = "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"n`"}"
Invoke-RestMethod -Uri "$User/api/v1/auth/register" -Method Post -Body $reg -ContentType "application/json; charset=utf-8" | Out-Null
$login = Invoke-RestMethod -Uri "$User/api/v1/auth/login" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`"}" -ContentType "application/json; charset=utf-8"
$headers = @{ Authorization = "Bearer $($login.access_token)"; "Idempotency-Key" = "nats-oc-1" }

$sku = "01NATSORDERCREATED00000001" # CHAR(26)
try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories" -Method Post -Body "{`"sku_id`":`"$sku`",`"available`":20}" -ContentType "application/json; charset=utf-8" | Out-Null
} catch { }

$itemBody = "{`"items`":[{`"sku_id`":`"$sku`",`"product_id`":`"01ARZ3NDEKTSV4RRFFQ69G5FAW`",`"title`":`"NATS`",`"price_cent`":100,`"quantity`":1}]}"
$created = Invoke-RestMethod -Uri "$Ord/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers $headers
$orderID = [string]$created.data.id
Assert "order created" ($orderID.Length -eq 26) $orderID
$eventID = "${orderID}:order_created"
Write-Host "order_id=$orderID event_id=$eventID"

# Inbox should be empty / unreachable while notification is down.
try {
    Invoke-RestMethod -Uri "$Notif/ping" -Method Get -TimeoutSec 1 | Out-Null
    Assert "notification down" $false "still reachable"
} catch {
    Assert "notification down" $true ""
}

Start-Process -FilePath ".\notification.exe" -WorkingDirectory $Root -WindowStyle Hidden
$ready = $false
for ($i = 0; $i -lt 20; $i++) {
    try {
        $p = Invoke-RestMethod -Uri "$Notif/ping" -Method Get -TimeoutSec 1
        if ($p.db -eq "ok") { $ready = $true; break }
    } catch { }
    Start-Sleep -Seconds 1
}
Assert "notification up" $ready ""

Start-Sleep -Seconds 2
$mine = Invoke-RestMethod -Uri "$Notif/api/v1/notifications/mine" -Method Get -Headers @{ Authorization = "Bearer $($login.access_token)" }
$hit = $false
# Single-element JSON arrays are unwrapped by Invoke-RestMethod; force array.
foreach ($row in @($mine.data)) {
    if ([string]$row.event_id -eq $eventID) { $hit = $true; break }
}
Assert "inbox has order_created from NATS backlog" $hit ("want=$eventID data=" + ($mine | ConvertTo-Json -Compress -Depth 6))

if ($Fail -gt 0) {
    Write-Host "FAILED count: $Fail" -ForegroundColor Red
    exit 1
}
Write-Host "NATS order_created backlog PASSED" -ForegroundColor Green
