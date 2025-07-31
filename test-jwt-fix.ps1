Write-Host "Testing JWT Fix for KrakenD Authentication" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

# Wait for services to be ready
Write-Host "Waiting for services to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Test 1: Login and get token
Write-Host "1. Testing login to get JWT token..." -ForegroundColor Cyan
$loginBody = @{
    email = "test@example.com"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8000/login" -Method POST -Body $loginBody -ContentType "application/json"

    if ($loginResponse.token) {
        Write-Host "Success: Login successful - token received" -ForegroundColor Green
        $token = $loginResponse.token
        Write-Host "Token (first 50 chars): $($token.Substring(0, [Math]::Min(50, $token.Length)))..." -ForegroundColor Gray

        Write-Host ""
        Write-Host "2. Testing authenticated endpoint..." -ForegroundColor Cyan

        $headers = @{
            "Authorization" = "Bearer $token"
        }

        try {
            $authResponse = Invoke-RestMethod -Uri "http://localhost:8000/users" -Method GET -Headers $headers
            Write-Host "SUCCESS: Authenticated request successful!" -ForegroundColor Green
            Write-Host "JWT validation is now working correctly" -ForegroundColor Green
            Write-Host "Response: $($authResponse | ConvertTo-Json -Depth 2)" -ForegroundColor Gray
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "FAILED: Authenticated request failed with code: $statusCode" -ForegroundColor Red
            Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    else {
        Write-Host "Login failed - no token received" -ForegroundColor Red
        Write-Host "Response: $($loginResponse | ConvertTo-Json)" -ForegroundColor Gray
    }
}
catch {
    Write-Host "Login failed" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Test completed." -ForegroundColor Green
