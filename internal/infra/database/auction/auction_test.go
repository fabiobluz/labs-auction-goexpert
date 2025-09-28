package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAutoCloseAuction(t *testing.T) {
	// Configura variável de ambiente para teste
	os.Setenv("AUCTION_INTERVAL", "1s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Conecta ao MongoDB de teste
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB não disponível para teste")
		return
	}
	defer client.Disconnect(ctx)

	database := client.Database("test_auction_db")
	defer database.Drop(ctx)

	// Cria repositório
	repo := NewAuctionRepository(database)

	// Cria um leilão
	auction, err := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"Test Description",
		auction_entity.New,
	)
	if err != nil {
		t.Fatalf("Erro ao criar leilão: %v", err)
	}

	// Salva o leilão
	err = repo.CreateAuction(ctx, auction)
	if err != nil {
		t.Fatalf("Erro ao salvar leilão: %v", err)
	}

	// Verifica se o leilão está ativo
	foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão: %v", err)
	}

	if foundAuction.Status != auction_entity.Active {
		t.Errorf("Leilão deveria estar ativo, mas status é: %v", foundAuction.Status)
	}

	// Aguarda o tempo de expiração + um pouco mais
	time.Sleep(2 * time.Second)

	// Força a verificação de leilões expirados
	repo.checkAndCloseExpiredAuctions()

	// Verifica se o leilão foi fechado automaticamente
	foundAuction, err = repo.FindAuctionById(ctx, auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão após fechamento: %v", err)
	}

	if foundAuction.Status != auction_entity.Completed {
		t.Errorf("Leilão deveria estar fechado, mas status é: %v", foundAuction.Status)
	}
}

func TestAuctionIntervalFromEnv(t *testing.T) {
	// Testa com intervalo customizado
	os.Setenv("AUCTION_INTERVAL", "30s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	interval := getAuctionInterval()
	expected := 30 * time.Second

	if interval != expected {
		t.Errorf("Intervalo esperado: %v, obtido: %v", expected, interval)
	}
}

func TestCheckIntervalFromEnv(t *testing.T) {
	// Testa com intervalo de verificação customizado
	os.Setenv("AUCTION_CHECK_INTERVAL", "5s")
	defer os.Unsetenv("AUCTION_CHECK_INTERVAL")

	interval := getCheckInterval()
	expected := 5 * time.Second

	if interval != expected {
		t.Errorf("Intervalo de verificação esperado: %v, obtido: %v", expected, interval)
	}
}

func TestContextTimeoutFromEnv(t *testing.T) {
	// Testa com timeout customizado
	os.Setenv("AUCTION_CONTEXT_TIMEOUT", "15s")
	defer os.Unsetenv("AUCTION_CONTEXT_TIMEOUT")

	timeout := getContextTimeout()
	expected := 15 * time.Second

	if timeout != expected {
		t.Errorf("Timeout esperado: %v, obtido: %v", expected, timeout)
	}
}

func TestAuctionIntervalDefault(t *testing.T) {
	// Testa com valor padrão quando variável não está definida
	os.Unsetenv("AUCTION_INTERVAL")

	interval := getAuctionInterval()
	expected := 5 * time.Minute

	if interval != expected {
		t.Errorf("Intervalo padrão esperado: %v, obtido: %v", expected, interval)
	}
}

func TestCheckIntervalDefault(t *testing.T) {
	// Testa com valor padrão quando variável não está definida
	os.Unsetenv("AUCTION_CHECK_INTERVAL")

	interval := getCheckInterval()
	expected := 10 * time.Second

	if interval != expected {
		t.Errorf("Intervalo de verificação padrão esperado: %v, obtido: %v", expected, interval)
	}
}

func TestContextTimeoutDefault(t *testing.T) {
	// Testa com valor padrão quando variável não está definida
	os.Unsetenv("AUCTION_CONTEXT_TIMEOUT")

	timeout := getContextTimeout()
	expected := 30 * time.Second

	if timeout != expected {
		t.Errorf("Timeout padrão esperado: %v, obtido: %v", expected, timeout)
	}
}

func TestUpdateAuctionStatus(t *testing.T) {
	// Conecta ao MongoDB de teste
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Skip("MongoDB não disponível para teste")
		return
	}
	defer client.Disconnect(ctx)

	database := client.Database("test_auction_db")
	defer database.Drop(ctx)

	// Cria repositório
	repo := NewAuctionRepository(database)

	// Cria um leilão
	auction, err := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"Test Description",
		auction_entity.New,
	)
	if err != nil {
		t.Fatalf("Erro ao criar leilão: %v", err)
	}

	// Salva o leilão
	err = repo.CreateAuction(ctx, auction)
	if err != nil {
		t.Fatalf("Erro ao salvar leilão: %v", err)
	}

	// Atualiza status para Completed
	err = repo.UpdateAuctionStatus(ctx, auction.Id, auction_entity.Completed)
	if err != nil {
		t.Fatalf("Erro ao atualizar status do leilão: %v", err)
	}

	// Verifica se o status foi atualizado
	foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão: %v", err)
	}

	if foundAuction.Status != auction_entity.Completed {
		t.Errorf("Status do leilão deveria ser Completed, mas é: %v", foundAuction.Status)
	}
}
