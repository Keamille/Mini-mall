# Cancel order: user :8081, inventory :8088, order :8089, gateway :8080
$User = "http://127.0.0.1:8081"
$Inv = "http://127.0.0.1:8088"
$Gw = "http://127.0.0.1:8080"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

$email = "cancel-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
$regBody = "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"c`"}"
Invoke-RestMethod -Uri "$User/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json; charset=utf-8" | Out-Null
$login = Invoke-RestMethod -Uri "$User/api/v1/auth/login" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`"}" -ContentType "application/json; charset=utf-8"
$headers = @{ Authorization = "Bearer $($login.access_token)"; "Idempotency-Key" = "cancel-1" }

$sku = "01CANCELORDERSKU0000000001"
try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories" -Method Post -Body "{`"sku_id`":`"$sku`",`"available`":10}" -ContentType "application/json; charset=utf-8" | Out-Null
} catch { }
$before = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
$beforeAvail = [int]$before.data.available

$itemBody = "{`"items`":[{`"sku_id`":`"$sku`",`"product_id`":`"01ARZ3NDEKTSV4RRFFQ69G5FAW`",`"title`":`"Phone`",`"price_cent`":100,`"quantity`":2}]}"
$created = Invoke-RestMethod -Uri "$Gw/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers $headers
Assert "created status 10" ($created.data.status -eq 10) ($created | ConvertTo-Json -Compress)
$orderID = $created.data.id

$mid = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "reserved after create" ([int]$mid.data.available -eq ($beforeAvail - 2)) ($mid.data.available)

$cancelUri = "$Gw/api/v1/orders/$orderID/cancel"
$cancelled = Invoke-RestMethod -Uri $cancelUri -Method Post -Headers @{ Authorization = $headers.Authorization }
Assert "cancelled status 30" ($cancelled.data.status -eq 30) ($cancelled | ConvertTo-Json -Compress)

$after = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "available restored" ([int]$after.data.available -eq $beforeAvail) ($after.data.available)
Assert "reserved 0" ([int]$after.data.reserved -eq 0) ($after.data.reserved)

try {
    Invoke-RestMethod -Uri $cancelUri -Method Post -Headers @{ Authorization = $headers.Authorization } | Out-Null
    Assert "second cancel 409" $false "expected 409"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert "second cancel 409" ($code -eq 409) "status=$code"
}
$after2 = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "stock stable" ([int]$after2.data.available -eq $beforeAvail) ($after2.data.available)

if ($Fail -eq 0) { Write-Host "order cancel PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
