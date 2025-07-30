#!/usr/bin/env pwsh

# ScaleBit Services Startup Script
# This script ensures proper startup order and handles common issues

Write-Host "Starting ScaleBit Platform Services..." -ForegroundColor Green

# Check if Docker is running
Write-Host "Checking Docker status..." -ForegroundColor Yellow
try {
    docker version | Out-Null
    Write-Host "Docker is running." -ForegroundColor Green
} catch {
    Write-Host "Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}

# Stop any existing containers
Write-Host "Stopping existing containers..." -ForegroundColor Yellow
docker-compose down --remove-orphans

# Clean up any dangling images/containers
Write-Host "Cleaning up..." -ForegroundColor Yellow
docker system prune -f

# Build and start services in proper order
Write-Host "Starting database..." -ForegroundColor Yellow
docker-compose up -d postgres

# Wait for database to be ready
Write-Host "Waiting for database to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Start backend services
Write-Host "Starting backend services..." -ForegroundColor Yellow
docker-compose up -d user-service product-service order-service payment-service

# Wait for backend services to be ready
Write-Host "Waiting for backend services to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Start KrakenD
Write-Host "Starting API Gateway (KrakenD)..." -ForegroundColor Yellow
docker-compose up -d krakend

# Wait for KrakenD to be ready
Write-Host "Waiting for API Gateway to be ready..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Start frontend
Write-Host "Starting frontend..." -ForegroundColor Yellow
docker-compose up -d frontend

Write-Host "All services started!" -ForegroundColor Green
Write-Host "Frontend available at: http://localhost" -ForegroundColor Cyan
Write-Host "API Gateway available at: http://localhost:8000" -ForegroundColor Cyan
Write-Host "Database available at: localhost:5432" -ForegroundColor Cyan

Write-Host "`nChecking service status..." -ForegroundColor Yellow
docker-compose ps

Write-Host "`nTo view logs, use: docker-compose logs -f [service-name]" -ForegroundColor Cyan
Write-Host "To stop all services, use: docker-compose down" -ForegroundColor Cyan
