# Inventory acceptance: product :8082 (for sku), inventory :8088
$Inv = "http://127.0.0.1:8088"
$Prod = "http://127.0.0.1:8082"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

Write-Host "=== ping ===" -ForegroundColor Cyan
try {
    $ping = Invoke-RestMethod -Uri "$Inv/ping" -Method Get
    Assert "inventory ping" ($ping.db -eq "ok") ($ping | ConvertTo-Json -Compress)
} catch { Assert "inventory ping" $false $_.Exception.Message; exit 1 }

Write-Host "`n=== get sku from product (or create) ===" -ForegroundColor Cyan
$skuId = $null
try {
    $prods = Invoke-RestMethod -Uri "$Prod/api/v1/products?all=1" -Method Get
    foreach ($p in $prods.data) {
        $detail = Invoke-RestMethod -Uri "$Prod/api/v1/products/$($p.id)" -Method Get
        if ($detail.data.skus -and $detail.data.skus.Count -gt 0) {
            $skuId = $detail.data.skus[0].id
            break
        }
    }
} catch { }

if (-not $skuId) {
    $cat = Invoke-RestMethod -Uri "$Prod/api/v1/categories" -Method Post -Body '{"name":"InvTest","sort":1}' -ContentType "application/json; charset=utf-8"
    $catId = $cat.data.id
    $pBody = "{`"title`":`"Inv Product`",`"category_id`":`"$catId`",`"status`":1}"
    $p = Invoke-RestMethod -Uri "$Prod/api/v1/products" -Method Post -Body $pBody -ContentType "application/json; charset=utf-8"
    $skuCode = "INV-" + (Get-Date -Format "HHmmss")
    $sBody = "{`"sku_code`":`"$skuCode`",`"price_cent`":100}"
    $sku = Invoke-RestMethod -Uri "$Prod/api/v1/products/$($p.data.id)/skus" -Method Post -Body $sBody -ContentType "application/json; charset=utf-8"
    $skuId = $sku.data.id
}
Assert "sku_id" ($skuId.Length -eq 26) "sku=$skuId"

Write-Host "`n=== init + get stock ===" -ForegroundColor Cyan
$initBody = "{`"sku_id`":`"$skuId`",`"available`":100}"
$init = Invoke-RestMethod -Uri "$Inv/api/v1/inventories" -Method Post -Body $initBody -ContentType "application/json; charset=utf-8"
Assert "POST init code=0" ($init.code -eq 0) ($init | ConvertTo-Json -Compress)
Assert "init available=100" ($init.data.available -eq 100) ($init.data.available)

$got = Invoke-RestMethod -Uri "$Inv/api/v1/inventories/$skuId" -Method Get
Assert "GET stock code=0" ($got.code -eq 0) ($got | ConvertTo-Json -Compress)
Assert "available=100" ($got.data.available -eq 100) ($got.data.available)
Assert "reserved=0" ($got.data.reserved -eq 0) ($got.data.reserved)

try {
    Invoke-RestMethod -Uri "$Inv/api/v1/inventories/00000000000000000000000000" -Method Get | Out-Null
    Assert "404 missing stock" $false "expected 404"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert "404 missing stock" ($code -eq 404) "status=$code"
}

Write-Host "`n=== done ===" -ForegroundColor Cyan
if ($Fail -eq 0) { Write-Host "inventory init/get PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
