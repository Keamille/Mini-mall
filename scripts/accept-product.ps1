# M1 product acceptance: docker mysql up, then go run ./cmd/product
$Base = "http://127.0.0.1:8082"
$Fail = 0

function Assert($name, $cond, $detail) {
    if ($cond) { Write-Host "[OK] $name" -ForegroundColor Green }
    else { Write-Host "[FAIL] $name -- $detail" -ForegroundColor Red; $script:Fail++ }
}

Write-Host "=== ping ===" -ForegroundColor Cyan
try {
    $ping = Invoke-RestMethod -Uri "$Base/ping" -Method Get
    Assert "ping" ($ping.db -eq "ok") ($ping | ConvertTo-Json -Compress)
} catch { Assert "ping" $false $_.Exception.Message; exit 1 }

Write-Host "`n=== category ===" -ForegroundColor Cyan
$catBody = '{"name":"Digital","sort":1}'
$cat = Invoke-RestMethod -Uri "$Base/api/v1/categories" -Method Post -Body $catBody -ContentType "application/json; charset=utf-8"
Assert "POST categories" ($cat.code -eq 0) ($cat | ConvertTo-Json -Compress)
$catId = $cat.data.id
Assert "category id len 26" ($catId.Length -eq 26) "id=$catId"

$catList = Invoke-RestMethod -Uri "$Base/api/v1/categories" -Method Get
Assert "GET categories" ($catList.data.Count -ge 1) "empty"

Write-Host "`n=== product ===" -ForegroundColor Cyan
$prodBody = "{`"title`":`"Accept Product`",`"category_id`":`"$catId`",`"status`":1}"
$prod = Invoke-RestMethod -Uri "$Base/api/v1/products" -Method Post -Body $prodBody -ContentType "application/json; charset=utf-8"
Assert "POST products" ($prod.code -eq 0) ($prod | ConvertTo-Json -Compress)
$prodId = $prod.data.id

$listOnSale = Invoke-RestMethod -Uri "$Base/api/v1/products" -Method Get
$ids = @($listOnSale.data | ForEach-Object { $_.id })
Assert "GET products on-sale" ($ids -contains $prodId) "not in list"

Write-Host "`n=== sku ===" -ForegroundColor Cyan
$skuCode = "ACCEPT-" + (Get-Date -Format "HHmmss")
$skuBody = "{`"sku_code`":`"$skuCode`",`"price_cent`":599900,`"attrs`":{`"color`":`"black`"}}"
$skuUri = "$Base/api/v1/products/$prodId/skus"
$sku = Invoke-RestMethod -Uri $skuUri -Method Post -Body $skuBody -ContentType "application/json; charset=utf-8"
Assert "POST skus" ($sku.code -eq 0) ($sku | ConvertTo-Json -Compress)

Write-Host "`n=== detail ===" -ForegroundColor Cyan
$detailUri = "$Base/api/v1/products/$prodId"
$detail = Invoke-RestMethod -Uri $detailUri -Method Get
Assert "GET detail title" ($detail.data.product.title -eq "Accept Product") ($detail | ConvertTo-Json -Compress)
Assert "detail skus" ($detail.data.skus.Count -ge 1) "no skus"
Assert "price_cent" ($detail.data.skus[0].price_cent -eq 599900) ($detail.data.skus[0].price_cent)

try {
    Invoke-RestMethod -Uri "$Base/api/v1/products/00000000000000000000000000" -Method Get | Out-Null
    Assert "404" $false "expected 404"
} catch {
    $code = $_.Exception.Response.StatusCode.value__
    Assert "404" ($code -eq 404) "status=$code"
}

Write-Host "`n=== done ===" -ForegroundColor Cyan
if ($Fail -eq 0) { Write-Host "M1 product PASSED" -ForegroundColor Green; exit 0 }
Write-Host "FAILED count: $Fail" -ForegroundColor Red; exit 1
