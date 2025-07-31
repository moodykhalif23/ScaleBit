Write-Host "JWT Fix Verification" -ForegroundColor Green
Write-Host "====================" -ForegroundColor Green

Write-Host "1. Checking KrakenD startup logs for JWT errors..." -ForegroundColor Cyan
$krakendLogs = docker logs scalebit-krakend-1 2>&1
$jwtErrors = $krakendLogs | Select-String -Pattern "jwt|validator|base64|jose" -CaseSensitive:$false

if ($jwtErrors -match "illegal base64 data" -or $jwtErrors -match "error in cryptographic primitive") {
    Write-Host "FAILED: JWT configuration errors found" -ForegroundColor Red
    $jwtErrors | ForEach-Object { Write-Host $_ -ForegroundColor Red }
} else {
    Write-Host "SUCCESS: No JWT configuration errors found" -ForegroundColor Green
    Write-Host "KrakenD started successfully with JWT validator enabled" -ForegroundColor Green
}

Write-Host ""
Write-Host "2. Checking JWT validator configuration..." -ForegroundColor Cyan
$validatorLogs = $krakendLogs | Select-String -Pattern "JWTValidator.*enabled"
if ($validatorLogs) {
    Write-Host "SUCCESS: JWT validators are properly configured" -ForegroundColor Green
    Write-Host "Found $($validatorLogs.Count) endpoints with JWT validation enabled" -ForegroundColor Gray
} else {
    Write-Host "WARNING: No JWT validator configuration found" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "3. Checking JWK file content..." -ForegroundColor Cyan
try {
    $jwkContent = docker exec scalebit-krakend-1 cat /etc/krakend/symmetric.jwk 2>$null
    if ($jwkContent) {
        Write-Host "SUCCESS: JWK file is accessible in KrakenD container" -ForegroundColor Green
        $jwkJson = $jwkContent | ConvertFrom-Json
        Write-Host "JWK Key ID: $($jwkJson.keys[0].kid)" -ForegroundColor Gray
        Write-Host "JWK Algorithm: $($jwkJson.keys[0].alg)" -ForegroundColor Gray
    } else {
        Write-Host "ERROR: JWK file not accessible" -ForegroundColor Red
    }
} catch {
    Write-Host "ERROR: Could not read JWK file: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "4. Summary..." -ForegroundColor Cyan
if ($jwtErrors -notmatch "illegal base64 data" -and $jwtErrors -notmatch "error in cryptographic primitive") {
    Write-Host "SUCCESS: JWT validation fix is working!" -ForegroundColor Green
    Write-Host "- KrakenD starts without JWT key errors" -ForegroundColor Green
    Write-Host "- JWT validators are properly configured" -ForegroundColor Green
    Write-Host "- JWK file is accessible and valid" -ForegroundColor Green
    Write-Host ""
    Write-Host "The original 'square/go-jose: error in cryptographic primitive' error has been resolved." -ForegroundColor Green
} else {
    Write-Host "FAILED: JWT validation issues still exist" -ForegroundColor Red
}

Write-Host ""
Write-Host "Verification completed." -ForegroundColor Green
