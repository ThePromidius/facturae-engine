param(
    [Parameter(Mandatory=$false)]
    [Switch]$Serve
)

if (-not (Test-Path ".venv")) {
    Write-Error "Virtual environment not found. Please run .\setup_docs.ps1 first."
    exit 1
}

# Activate venv
. ".\.venv\Scripts\Activate.ps1"

if ($Serve) {
    Write-Host "Starting live documentation server at http://localhost:8000..." -ForegroundColor Cyan
    sphinx-autobuild docs docs/_build/html
} else {
    Write-Host "Building documentation..." -ForegroundColor Cyan
    sphinx-build -b html docs docs/_build/html
    Write-Host "Build complete. Output in docs/_build/html" -ForegroundColor Green
}
