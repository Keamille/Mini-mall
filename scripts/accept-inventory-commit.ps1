# Commit stock: user :8081, inventory :8088, order :8089
$User = "http://127.0.0.1:8081"
$Inv = "http://127.0.0.1:8088"
$Ord = "http://127.0.0.1:8089"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

$email = "commit-" + (Get-Date -Format "yyyyMMddHHmmss") + "@test.local"
$regBody = "{`"email`":`"$email`",`"password`":`"Pass123456`",`"re_password`":`"Pass123456`",`"nickname`":`"c`"}"
Invoke-RestMethod -Uri "$User/api/v1/auth/register" -Method Post -Body $regBody -ContentType "application/json; charset=utf-8" | Out-Null
$login = Invoke-RestMethod -Uri "$User/api/v1/auth/login" -Method Post -Body "{`"email`":`"$email`",`"password`":`"Pass123456`"}" -ContentType "application/json; charset=utf-8"
$headers = @{ Authorization = "Bearer $($login.access_token)" }

$sku = "01COMMITSTOCKSKU0000000001"
try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories" -Method Post -Body "{`"sku_id`":`"$sku`",`"available`":20}" -ContentType "application/json; charset=utf-8" | Out-Null
} catch { }
$before = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
$beforeAvail = [int]$before.data.available

$itemBody = "{`"items`":[{`"sku_id`":`"$sku`",`"product_id`":`"01ARZ3NDEKTSV4RRFFQ69G5FAW`",`"title`":`"Phone`",`"price_cent`":100,`"quantity`":2}]}"
$h1 = $headers.Clone()
$h1["Idempotency-Key"] = "commit-1"
$created = Invoke-RestMethod -Uri "$Ord/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers $h1
$orderID = $created.data.id
Assert "order pending" ($created.data.status -eq 10) ($created | ConvertTo-Json -Compress)

$mid = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "available down" ([int]$mid.data.available -eq ($beforeAvail - 2)) ($mid.data.available)
Assert "reserved 2" ([int]$mid.data.reserved -ge 2) ($mid.data.reserved)

$commitBody = "{`"order_id`":`"$orderID`"}"
Invoke-RestMethod -Uri "$Inv/api/v1/inventories/commit" -Method Post -Body $commitBody -ContentType "application/json; charset=utf-8" | Out-Null
$after = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "available unchanged" ([int]$after.data.available -eq ($beforeAvail - 2)) ($after.data.available)
$reservedAfter = [int]$after.data.reserved
Assert "reserved decreased" ($reservedAfter -eq ([int]$mid.data.reserved - 2)) "$reservedAfter vs $($mid.data.reserved)"

Invoke-RestMethod -Uri "$Inv/api/v1/inventories/commit" -Method Post -Body $commitBody -ContentType "application/json; charset=utf-8" | Out-Null
$again = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$sku" -Method Get
Assert "second commit stable" (([int]$again.data.available -eq [int]$after.data.available) -and ([int]$again.data.reserved -eq $reservedAfter)) ($again.data | ConvertTo-Json -Compress)

$h2 = $headers.Clone()
$h2["Idempotency-Key"] = "commit-2"
$created2 = Invoke-RestMethod -Uri "$Ord/api/v1/orders" -Method Post -Body $itemBody -ContentType "application/json; charset=utf-8" -Headers $h2
Invoke-RestMethod -Uri "$Ord/api/v1/orders/$($created2.data.id)/cancel" -Method Post -Headers $headers | Out-Null
try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories/commit" -Method Post -Body "{`"order_id`":`"$($created2.data.id)`"}" -ContentType "application/json; charset=utf-8" | Out-Null
    Assert "commit cancelled fails" $false "expected error"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert "commit cancelled fails" ($code -eq 400 -or $code -eq 409) "status=$code"
}

if ($Fail -eq 0) { Write-Host "inventory commit PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
