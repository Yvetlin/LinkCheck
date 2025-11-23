# Simple test script for API
# Run: powershell -ExecutionPolicy Bypass -File tests/test.ps1

$baseUrl = "http://localhost:8080"
$reportsDir = "reports"

# Create reports directory if it doesn't exist
if (-not (Test-Path $reportsDir)) {
    New-Item -ItemType Directory -Path $reportsDir | Out-Null
}

Write-Host "Test 1: Check links" -ForegroundColor Green
$body1 = '{"links": ["google.com", "yandex.ru", "malformedlink.gg"]}'

try {
    $response1 = Invoke-RestMethod -Uri "$baseUrl/submit" -Method POST -ContentType "application/json" -Body $body1
    Write-Host "Success! Response:" -ForegroundColor Green
    $response1 | ConvertTo-Json -Depth 10
    $linksNum1 = $response1.links_num
    Write-Host "Links number: $linksNum1" -ForegroundColor Yellow
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`nTest 2: Check another set of links" -ForegroundColor Green
$body2 = '{"links": ["gg.c", "yandex.ru"]}'

try {
    $response2 = Invoke-RestMethod -Uri "$baseUrl/submit" -Method POST -ContentType "application/json" -Body $body2
    Write-Host "Success! Response:" -ForegroundColor Green
    $response2 | ConvertTo-Json -Depth 10
    $linksNum2 = $response2.links_num
    Write-Host "Links number: $linksNum2" -ForegroundColor Yellow
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`nTest 3: Get PDF report" -ForegroundColor Green
$body3 = "{`"links_list`": [$linksNum1, $linksNum2]}"

try {
    $pdfPath = Join-Path $reportsDir "report_$linksNum1`_$linksNum2.pdf"
    Invoke-RestMethod -Uri "$baseUrl/report" -Method POST -ContentType "application/json" -Body $body3 -OutFile $pdfPath
    Write-Host "PDF report saved to $pdfPath" -ForegroundColor Green
    $fileSize = (Get-Item $pdfPath).Length
    Write-Host "File size: $fileSize bytes" -ForegroundColor Yellow
} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`nAll tests passed!" -ForegroundColor Green
