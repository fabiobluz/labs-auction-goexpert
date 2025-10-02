#!/bin/bash

# Script para executar testes Docker
# Uso: ./scripts/test-docker.sh [test-type]
# Tipos: all, auction, integration, auto-close, performance

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para imprimir mensagens coloridas
print_message() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Função para limpar containers
cleanup() {
    print_message "Limpando containers e volumes..."
    docker-compose -f docker-compose.test.yml down -v --remove-orphans
    docker system prune -f
}

# Função para executar testes
run_tests() {
    local test_type=$1
    
    print_message "Iniciando testes Docker: $test_type"
    
    case $test_type in
        "all")
            print_message "Executando todos os testes..."
            docker-compose -f docker-compose.test.yml up --build test-all
            ;;
        "auction")
            print_message "Executando testes de auction..."
            docker-compose -f docker-compose.test.yml up --build test-auction
            ;;
        "integration")
            print_message "Executando testes de integração..."
            docker-compose -f docker-compose.test.yml up --build test-integration
            ;;
        "auto-close")
            print_message "Executando testes de fechamento automático..."
            docker-compose -f docker-compose.test.yml up --build test-auto-close
            ;;
        "performance")
            print_message "Executando testes de performance..."
            docker-compose -f docker-compose.test.yml up --build test-performance
            ;;
        *)
            print_error "Tipo de teste inválido: $test_type"
            print_message "Tipos disponíveis: all, auction, integration, auto-close, performance"
            exit 1
            ;;
    esac
}

# Função para mostrar ajuda
show_help() {
    echo "Uso: $0 [test-type] [options]"
    echo ""
    echo "Tipos de teste:"
    echo "  all         - Executa todos os testes"
    echo "  auction     - Executa testes específicos do auction"
    echo "  integration - Executa testes de integração"
    echo "  auto-close  - Executa testes de fechamento automático"
    echo "  performance - Executa testes de performance"
    echo ""
    echo "Opções:"
    echo "  --clean     - Limpa containers e volumes antes de executar"
    echo "  --help      - Mostra esta ajuda"
    echo ""
    echo "Exemplos:"
    echo "  $0 all"
    echo "  $0 auction --clean"
    echo "  $0 auto-close"
}

# Função principal
main() {
    local test_type="all"
    local clean=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --clean)
                clean=true
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                test_type=$1
                shift
                ;;
        esac
    done
    
    # Verificar se Docker está rodando
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker não está rodando. Por favor, inicie o Docker Desktop."
        exit 1
    fi
    
    # Limpar se solicitado
    if [ "$clean" = true ]; then
        cleanup
    fi
    
    # Executar testes
    run_tests "$test_type"
    
    # Limpar após execução
    cleanup
    
    print_success "Testes concluídos!"
}

# Executar função principal
main "$@"
