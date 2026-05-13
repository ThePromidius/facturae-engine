# Setup script for documentation environment
Write-Host "Setting up documentation environment..." -ForegroundColor Cyan

if (-not (Test-Path ".venv")) {
    Write-Host "Creating virtual environment..."
    python -m venv .venv
}

Write-Host "Activating venv and installing dependencies..."
& ".\.venv\Scripts\Activate.ps1"
python -m pip install --upgrade pip
python -m pip install -r requirements-docs.txt

Write-Host "Environment ready! Use .\make_docs.ps1 to build or serve documentation." -ForegroundColor Green
