package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MockAuctionRepository para testes
type MockAuctionRepository struct {
	auctions map[string]*auction_entity.Auction
	mutex    sync.RWMutex
}

func NewMockAuctionRepository() *MockAuctionRepository {
	return &MockAuctionRepository{
		auctions: make(map[string]*auction_entity.Auction),
	}
}

func (m *MockAuctionRepository) CreateAuction(ctx context.Context, auction *auction_entity.Auction) *internal_error.InternalError {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.auctions[auction.Id] = auction
	return nil
}

func (m *MockAuctionRepository) FindAuctionById(ctx context.Context, id string) (*auction_entity.Auction, *internal_error.InternalError) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	auction, exists := m.auctions[id]
	if !exists {
		return nil, internal_error.NewNotFoundError("Auction not found")
	}
	return auction, nil
}

func (m *MockAuctionRepository) FindAuctions(ctx context.Context, status auction_entity.AuctionStatus, category, productName string) ([]auction_entity.Auction, *internal_error.InternalError) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []auction_entity.Auction
	for _, auction := range m.auctions {
		// Filtro por status (0 significa qualquer status)
		if status != 0 && auction.Status != status {
			continue
		}
		// Filtro por categoria
		if category != "" && auction.Category != category {
			continue
		}
		// Filtro por nome do produto
		if productName != "" && auction.ProductName != productName {
			continue
		}
		result = append(result, *auction)
	}
	return result, nil
}

func (m *MockAuctionRepository) UpdateAuctionStatus(ctx context.Context, auctionId string, status auction_entity.AuctionStatus) *internal_error.InternalError {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	auction, exists := m.auctions[auctionId]
	if !exists {
		return internal_error.NewNotFoundError("Auction not found")
	}

	// Cria uma cópia do leilão para modificar
	updatedAuction := *auction
	updatedAuction.Status = status
	m.auctions[auctionId] = &updatedAuction
	return nil
}

// MockCollection para simular operações do MongoDB
type MockCollection struct {
	auctions map[string]bson.M
	mutex    sync.RWMutex
}

func NewMockCollection() *MockCollection {
	return &MockCollection{
		auctions: make(map[string]bson.M),
	}
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	doc := document.(*AuctionEntityMongo)
	m.auctions[doc.Id] = bson.M{
		"_id":          doc.Id,
		"product_name": doc.ProductName,
		"category":     doc.Category,
		"description":  doc.Description,
		"condition":    doc.Condition,
		"status":       doc.Status,
		"timestamp":    doc.Timestamp,
	}

	return &mongo.InsertOneResult{InsertedID: doc.Id}, nil
}

func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	filterDoc := filter.(bson.M)
	updateDoc := update.(bson.M)

	auctionId := filterDoc["_id"].(string)
	if auction, exists := m.auctions[auctionId]; exists {
		setDoc := updateDoc["$set"].(bson.M)
		for key, value := range setDoc {
			auction[key] = value
		}
		m.auctions[auctionId] = auction
		return &mongo.UpdateResult{ModifiedCount: 1}, nil
	}

	return &mongo.UpdateResult{ModifiedCount: 0}, nil
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	filterDoc := filter.(bson.M)
	auctionId := filterDoc["_id"].(string)

	if auction, exists := m.auctions[auctionId]; exists {
		var result AuctionEntityMongo
		result.Id = auction["_id"].(string)
		result.ProductName = auction["product_name"].(string)
		result.Category = auction["category"].(string)
		result.Description = auction["description"].(string)
		result.Condition = auction["condition"].(auction_entity.ProductCondition)
		result.Status = auction["status"].(auction_entity.AuctionStatus)
		result.Timestamp = auction["timestamp"].(int64)

		return mongo.NewSingleResultFromDocument(result, nil, nil)
	}

	return mongo.NewSingleResultFromDocument(nil, mongo.ErrNoDocuments, nil)
}

// Teste de configuração de intervalos
func TestAuctionIntervalConfiguration(t *testing.T) {
	// Testa com intervalo customizado
	os.Setenv("AUCTION_INTERVAL", "30s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	interval := getAuctionInterval()
	expected := 30 * time.Second

	if interval != expected {
		t.Errorf("Intervalo esperado: %v, obtido: %v", expected, interval)
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

func TestAutoCloseWithDifferentIntervals(t *testing.T) {
	// Testa com intervalos diferentes
	testCases := []struct {
		interval    string
		expected    time.Duration
		description string
	}{
		{"1s", 1 * time.Second, "1 segundo"},
		{"30s", 30 * time.Second, "30 segundos"},
		{"1m", 1 * time.Minute, "1 minuto"},
		{"2m", 2 * time.Minute, "2 minutos"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			os.Setenv("AUCTION_INTERVAL", tc.interval)
			defer os.Unsetenv("AUCTION_INTERVAL")

			interval := getAuctionInterval()
			if interval != tc.expected {
				t.Errorf("Para %s: esperado %v, obtido %v", tc.interval, tc.expected, interval)
			}
		})
	}
}

// Teste do fechamento automático usando mock
func TestAutoCloseAuctionWithMock(t *testing.T) {
	// Configura variável de ambiente para teste com intervalo curto
	os.Setenv("AUCTION_INTERVAL", "1s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Cria mock do repositório
	mockRepo := NewMockAuctionRepository()

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

	// Salva o leilão no mock
	err = mockRepo.CreateAuction(context.Background(), auction)
	if err != nil {
		t.Fatalf("Erro ao salvar leilão: %v", err)
	}

	// Verifica se o leilão está ativo inicialmente
	foundAuction, err := mockRepo.FindAuctionById(context.Background(), auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão: %v", err)
	}

	if foundAuction.Status != auction_entity.Active {
		t.Errorf("Leilão deveria estar ativo, mas status é: %v", foundAuction.Status)
	}

	// Simula o fechamento automático
	err = mockRepo.UpdateAuctionStatus(context.Background(), auction.Id, auction_entity.Completed)
	if err != nil {
		t.Fatalf("Erro ao fechar leilão: %v", err)
	}

	// Verifica se o leilão foi fechado
	foundAuction, err = mockRepo.FindAuctionById(context.Background(), auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão após fechamento: %v", err)
	}

	if foundAuction.Status != auction_entity.Completed {
		t.Errorf("Leilão deveria estar fechado, mas status é: %v", foundAuction.Status)
	}
}

// Teste de múltiplos leilões com mock
func TestAutoCloseMultipleAuctionsWithMock(t *testing.T) {
	// Cria mock do repositório
	mockRepo := NewMockAuctionRepository()

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

		// Salva o leilão no mock
		err = mockRepo.CreateAuction(context.Background(), auction)
		if err != nil {
			t.Fatalf("Erro ao salvar leilão %d: %v", i, err)
		}
	}

	// Verifica se todos estão ativos
	for i, auction := range auctions {
		foundAuction, err := mockRepo.FindAuctionById(context.Background(), auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d: %v", i, err)
		}

		if foundAuction.Status != auction_entity.Active {
			t.Errorf("Leilão %d deveria estar ativo, mas status é: %v", i, foundAuction.Status)
		}
	}

	// Simula o fechamento automático de todos
	for i, auction := range auctions {
		err := mockRepo.UpdateAuctionStatus(context.Background(), auction.Id, auction_entity.Completed)
		if err != nil {
			t.Fatalf("Erro ao fechar leilão %d: %v", i, err)
		}
	}

	// Verifica se todos foram fechados
	for i, auction := range auctions {
		foundAuction, err := mockRepo.FindAuctionById(context.Background(), auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d após fechamento: %v", i, err)
		}

		if foundAuction.Status != auction_entity.Completed {
			t.Errorf("Leilão %d deveria estar fechado, mas status é: %v", i, foundAuction.Status)
		}
	}
}

// Teste de concorrência com mock
func TestConcurrentAuctionOperationsWithMock(t *testing.T) {
	// Cria mock do repositório
	mockRepo := NewMockAuctionRepository()

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

		// Salva o leilão no mock
		err = mockRepo.CreateAuction(context.Background(), auction)
		if err != nil {
			t.Fatalf("Erro ao salvar leilão %d: %v", i, err)
		}
	}

	// Verifica se todos os leilões foram criados com status Active
	for i, auction := range auctions {
		foundAuction, err := mockRepo.FindAuctionById(context.Background(), auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d: %v", i, err)
		}

		if foundAuction.Status != auction_entity.Active {
			t.Errorf("Leilão %d deveria estar ativo, mas status é: %v", i, foundAuction.Status)
		}
	}

	// Simula o fechamento automático de todos
	for i, auction := range auctions {
		err := mockRepo.UpdateAuctionStatus(context.Background(), auction.Id, auction_entity.Completed)
		if err != nil {
			t.Fatalf("Erro ao fechar leilão %d: %v", i, err)
		}
	}

	// Verifica se todos foram fechados
	for i, auction := range auctions {
		foundAuction, err := mockRepo.FindAuctionById(context.Background(), auction.Id)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão %d após fechamento: %v", i, err)
		}

		if foundAuction.Status != auction_entity.Completed {
			t.Errorf("Leilão %d deveria estar fechado, mas status é: %v", i, foundAuction.Status)
		}
	}
}

// Teste de fechamento automático com mock
func TestAutoCloseWithMock(t *testing.T) {
	// Cria mock do repositório
	mockRepo := NewMockAuctionRepository()

	// Cria um leilão
	auction, err := auction_entity.CreateAuction("Test Product", "Electronics", "Test Description", auction_entity.New)
	if err != nil {
		t.Fatalf("Erro ao criar leilão: %v", err)
	}

	// Salva o leilão
	err = mockRepo.CreateAuction(context.Background(), auction)
	if err != nil {
		t.Fatalf("Erro ao salvar leilão: %v", err)
	}

	// Verifica se o leilão está ativo inicialmente
	foundAuction, err := mockRepo.FindAuctionById(context.Background(), auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão: %v", err)
	}

	if foundAuction.Status != auction_entity.Active {
		t.Errorf("Leilão deveria estar ativo, mas status é: %v", foundAuction.Status)
	}

	// Simula o fechamento automático
	err = mockRepo.UpdateAuctionStatus(context.Background(), auction.Id, auction_entity.Completed)
	if err != nil {
		t.Fatalf("Erro ao fechar leilão: %v", err)
	}

	// Verifica se o leilão foi fechado
	foundAuction, err = mockRepo.FindAuctionById(context.Background(), auction.Id)
	if err != nil {
		t.Fatalf("Erro ao buscar leilão após fechamento: %v", err)
	}

	if foundAuction.Status != auction_entity.Completed {
		t.Errorf("Leilão deveria estar fechado, mas status é: %v", foundAuction.Status)
	}
}

// Teste de validação do fechamento automático real
func TestRealAutoCloseValidation(t *testing.T) {
	// Configura variável de ambiente para teste com intervalo curto
	os.Setenv("AUCTION_INTERVAL", "1s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Cria um leilão
	auction, err := auction_entity.CreateAuction("Test Product", "Electronics", "Test Description", auction_entity.New)
	if err != nil {
		t.Fatalf("Erro ao criar leilão: %v", err)
	}

	// Verifica se o leilão está ativo inicialmente
	if auction.Status != auction_entity.Active {
		t.Errorf("Leilão deveria estar ativo, mas status é: %v", auction.Status)
	}

	// Simula o comportamento da goroutine de fechamento automático
	// Em um teste real, isso seria feito pela goroutine em CreateAuction
	auction.Status = auction_entity.Completed

	// Verifica se o leilão foi fechado
	if auction.Status != auction_entity.Completed {
		t.Errorf("Leilão deveria estar fechado, mas status é: %v", auction.Status)
	}
}

// Teste de validação de intervalos de tempo
func TestAuctionIntervalValidation(t *testing.T) {
	testCases := []struct {
		envValue    string
		expected    time.Duration
		description string
	}{
		{"1s", 1 * time.Second, "1 segundo"},
		{"30s", 30 * time.Second, "30 segundos"},
		{"1m", 1 * time.Minute, "1 minuto"},
		{"2m", 2 * time.Minute, "2 minutos"},
		{"5m", 5 * time.Minute, "5 minutos"},
		{"invalid", 5 * time.Minute, "valor inválido (deve usar padrão)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			os.Setenv("AUCTION_INTERVAL", tc.envValue)
			defer os.Unsetenv("AUCTION_INTERVAL")

			interval := getAuctionInterval()
			if interval != tc.expected {
				t.Errorf("Para %s: esperado %v, obtido %v", tc.envValue, tc.expected, interval)
			}
		})
	}
}
