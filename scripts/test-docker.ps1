# Script PowerShell para executar testes Docker
# Uso: .\scripts\test-docker.ps1 [test-type]
# Tipos: all, auction, integration, auto-close, performance

param(
    [string]$TestType = "all",
    [switch]$Clean,
    [switch]$Help
)

# Função para imprimir mensagens coloridas
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Função para limpar containers
function Cleanup {
    Write-Info "Limpando containers e volumes..."
    docker-compose -f docker-compose.test.yml down -v --remove-orphans
    docker system prune -f
}

# Função para executar testes
function Run-Tests {
    param([string]$TestType)
    
    Write-Info "Iniciando testes Docker: $TestType"
    
    switch ($TestType) {
        "all" {
            Write-Info "Executando todos os testes..."
            docker-compose -f docker-compose.test.yml up --build test-all
        }
        "auction" {
            Write-Info "Executando testes de auction..."
            docker-compose -f docker-compose.test.yml up --build test-auction
        }
        "integration" {
            Write-Info "Executando testes de integração..."
            docker-compose -f docker-compose.test.yml up --build test-integration
        }
        "auto-close" {
            Write-Info "Executando testes de fechamento automático..."
            docker-compose -f docker-compose.test.yml up --build test-auto-close
        }
        "performance" {
            Write-Info "Executando testes de performance..."
            docker-compose -f docker-compose.test.yml up --build test-performance
        }
        default {
            Write-Error "Tipo de teste inválido: $TestType"
            Write-Info "Tipos disponíveis: all, auction, integration, auto-close, performance"
            exit 1
        }
    }
}

# Função para mostrar ajuda
function Show-Help {
    Write-Host "Uso: .\scripts\test-docker.ps1 [test-type] [options]"
    Write-Host ""
    Write-Host "Tipos de teste:"
    Write-Host "  all         - Executa todos os testes"
    Write-Host "  auction     - Executa testes específicos do auction"
    Write-Host "  integration - Executa testes de integração"
    Write-Host "  auto-close  - Executa testes de fechamento automático"
    Write-Host "  performance - Executa testes de performance"
    Write-Host ""
    Write-Host "Opções:"
    Write-Host "  -Clean      - Limpa containers e volumes antes de executar"
    Write-Host "  -Help       - Mostra esta ajuda"
    Write-Host ""
    Write-Host "Exemplos:"
    Write-Host "  .\scripts\test-docker.ps1 all"
    Write-Host "  .\scripts\test-docker.ps1 auction -Clean"
    Write-Host "  .\scripts\test-docker.ps1 auto-close"
}

# Função principal
function Main {
    # Mostrar ajuda se solicitado
    if ($Help) {
        Show-Help
        return
    }
    
    # Verificar se Docker está rodando
    try {
        docker info | Out-Null
    }
    catch {
        Write-Error "Docker não está rodando. Por favor, inicie o Docker Desktop."
        exit 1
    }
    
    # Limpar se solicitado
    if ($Clean) {
        Cleanup
    }
    
    # Executar testes
    Run-Tests $TestType
    
    # Limpar após execução
    Cleanup
    
    Write-Success "Testes concluídos!"
}

# Executar função principal
Main
