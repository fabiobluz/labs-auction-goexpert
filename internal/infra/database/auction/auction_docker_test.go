//go:build integration
// +build integration

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

// Teste de fechamento automático com MongoDB real
func TestAutoCloseAuctionWithMongoDB(t *testing.T) {
	// Configura variável de ambiente para teste com intervalo curto
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Conecta ao MongoDB de teste
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-test:27017"))
	if err != nil {
		t.Skip("MongoDB não disponível para teste - use Docker Compose")
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

	// Verifica se o leilão está ativo inicialmente
	foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão: %v", err)
	}

	if foundAuction.Status != auction_entity.Active {
		t.Errorf("Leilão deveria estar ativo, mas status é: %v", foundAuction.Status)
	}

	// Aguarda o tempo de expiração + um pouco mais para garantir que a goroutine execute
	time.Sleep(3 * time.Second)

	// Verifica se o leilão foi fechado automaticamente
	foundAuction, err = repo.FindAuctionById(ctx, auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão após fechamento: %v", err)
	}

	if foundAuction.Status != auction_entity.Completed {
		t.Errorf("Leilão deveria estar fechado automaticamente, mas status é: %v", foundAuction.Status)
	}
}

// Teste de múltiplos leilões com fechamento automático
func TestAutoCloseMultipleAuctionsWithMongoDB(t *testing.T) {
	// Configura variável de ambiente para teste com intervalo curto
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Conecta ao MongoDB de teste
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-test:27017"))
	if err != nil {
		t.Skip("MongoDB não disponível para teste - use Docker Compose")
		return
	}
	defer client.Disconnect(ctx)

	database := client.Database("test_auction_db")
	defer database.Drop(ctx)

	// Cria repositório
	repo := NewAuctionRepository(database)

	// Cria múltiplos leilões
	auctions := make([]*auction_entity.Auction, 3)
	for i := 0; i < 3; i++ {
		auction, err := auction_entity.CreateAuction(
			"Test Product",
			"Electronics",
			"Test Description",
			auction_entity.New,
		)
		if err != nil {
			t.Fatalf("Erro ao criar leilão %d: %v", i, err)
		}
		auctions[i] = auction

		// Salva o leilão
		err = repo.CreateAuction(ctx, auction)
		if err != nil {
			t.Fatalf("Erro ao salvar leilão %d: %v", i, err)
		}
	}

	// Aguarda o tempo de expiração + um pouco mais
	time.Sleep(3 * time.Second)

	// Verifica se todos os leilões foram fechados automaticamente
	for i, auction := range auctions {
		foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d: %v", i, err)
		}

		if foundAuction.Status != auction_entity.Completed {
			t.Errorf("Leilão %d deveria ter sido fechado automaticamente, mas status é: %v", i, foundAuction.Status)
		}
	}
}

// Teste de concorrência com fechamento automático
func TestConcurrentAuctionCreationWithMongoDB(t *testing.T) {
	// Configura variável de ambiente para teste
	os.Setenv("AUCTION_INTERVAL", "3s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Conecta ao MongoDB de teste
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-test:27017"))
	if err != nil {
		t.Skip("MongoDB não disponível para teste - use Docker Compose")
		return
	}
	defer client.Disconnect(ctx)

	database := client.Database("test_auction_db")
	defer database.Drop(ctx)

	// Cria repositório
	repo := NewAuctionRepository(database)

	// Cria vários leilões rapidamente
	auctions := make([]*auction_entity.Auction, 5)
	for i := 0; i < 5; i++ {
		auction, err := auction_entity.CreateAuction(
			"Test Product",
			"Electronics",
			"Test Description",
			auction_entity.New,
		)
		if err != nil {
			t.Fatalf("Erro ao criar leilão %d: %v", i, err)
		}
		auctions[i] = auction

		// Salva o leilão
		err = repo.CreateAuction(ctx, auction)
		if err != nil {
			t.Fatalf("Erro ao salvar leilão %d: %v", i, err)
		}
	}

	// Verifica se todos os leilões foram criados com status Active
	for i, auction := range auctions {
		foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d: %v", i, err)
		}

		if foundAuction.Status != auction_entity.Active {
			t.Errorf("Leilão %d deveria estar ativo, mas status é: %v", i, foundAuction.Status)
		}
	}

	// Aguarda o tempo de expiração
	time.Sleep(4 * time.Second)

	// Verifica se todos foram fechados automaticamente
	for i, auction := range auctions {
		foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d após fechamento: %v", i, err)
		}

		if foundAuction.Status != auction_entity.Completed {
			t.Errorf("Leilão %d deveria ter sido fechado automaticamente, mas status é: %v", i, foundAuction.Status)
		}
	}
}

// Teste de performance com fechamento automático
func TestPerformanceWithAutoClose(t *testing.T) {
	// Configura variável de ambiente para teste
	os.Setenv("AUCTION_INTERVAL", "1s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Conecta ao MongoDB de teste
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-test:27017"))
	if err != nil {
		t.Skip("MongoDB não disponível para teste - use Docker Compose")
		return
	}
	defer client.Disconnect(ctx)

	database := client.Database("test_auction_db")
	defer database.Drop(ctx)

	// Cria repositório
	repo := NewAuctionRepository(database)

	// Mede o tempo de criação de múltiplos leilões
	start := time.Now()
	auctions := make([]*auction_entity.Auction, 10)

	for i := 0; i < 10; i++ {
		auction, err := auction_entity.CreateAuction(
			"Test Product",
			"Electronics",
			"Test Description",
			auction_entity.New,
		)
		if err != nil {
			t.Fatalf("Erro ao criar leilão %d: %v", i, err)
		}
		auctions[i] = auction

		// Salva o leilão
		err = repo.CreateAuction(ctx, auction)
		if err != nil {
			t.Fatalf("Erro ao salvar leilão %d: %v", i, err)
		}
	}

	creationTime := time.Since(start)
	t.Logf("Tempo para criar 10 leilões: %v", creationTime)

	// Aguarda o fechamento automático
	time.Sleep(2 * time.Second)

	// Verifica se todos foram fechados
	closedCount := 0
	for i, auction := range auctions {
		foundAuction, err := repo.FindAuctionById(ctx, auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d: %v", i, err)
		}

		if foundAuction.Status == auction_entity.Completed {
			closedCount++
		}
	}

	if closedCount != 10 {
		t.Errorf("Deveria ter 10 leilões fechados, mas encontrou %d", closedCount)
	}

	t.Logf("Performance: %d leilões criados e fechados em %v", len(auctions), time.Since(start))
}

// Teste de configuração de intervalos com MongoDB
func TestAuctionIntervalConfigurationWithMongoDB(t *testing.T) {
	testCases := []struct {
		interval    string
		description string
	}{
		{"1s", "1 segundo"},
		{"30s", "30 segundos"},
		{"1m", "1 minuto"},
		{"2m", "2 minutos"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Configura variável de ambiente
			os.Setenv("AUCTION_INTERVAL", tc.interval)
			defer os.Unsetenv("AUCTION_INTERVAL")

			// Conecta ao MongoDB de teste
			ctx := context.Background()
			client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb-test:27017"))
			if err != nil {
				t.Skip("MongoDB não disponível para teste - use Docker Compose")
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

			// Aguarda o fechamento automático
			time.Sleep(2 * time.Second)

			// Verifica se foi fechado
			foundAuction, err = repo.FindAuctionById(ctx, auction.Id)
			if err != nil {
				t.Fatalf("Erro ao buscar leilão após fechamento: %v", err)
			}

			if foundAuction.Status != auction_entity.Completed {
				t.Errorf("Leilão deveria estar fechado, mas status é: %v", foundAuction.Status)
			}
		})
	}
}
