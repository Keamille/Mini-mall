# Generate Go stubs for service-to-service protobuf contracts.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$protos = @(
  "api/proto/inventory/v1/inventory.proto",
  "api/proto/coupon/v1/coupon.proto",
  "api/proto/notification/v1/notification.proto",
  "api/proto/order/v1/order.proto",
  "api/proto/search/v1/search.proto"
)
foreach ($proto in $protos) {
  if (-not (Test-Path $proto)) {
    Write-Error "missing $proto"
  }
}

protoc -I api/proto `
  --go_out=. --go_opt=module=github.com/your-org/go-minimall `
  --go-grpc_out=. --go-grpc_opt=module=github.com/your-org/go-minimall `
  $protos

Write-Host "generated api/gen/*/*.pb.go"
