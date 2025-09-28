package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection          *mongo.Collection
	auctionInterval     time.Duration
	checkInterval       time.Duration
	contextTimeout      time.Duration
	auctionEndTimeMap   map[string]time.Time
	auctionEndTimeMutex *sync.Mutex
	stopChan            chan struct{}
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection:          database.Collection("auctions"),
		auctionInterval:     getAuctionInterval(),
		checkInterval:       getCheckInterval(),
		contextTimeout:      getContextTimeout(),
		auctionEndTimeMap:   make(map[string]time.Time),
		auctionEndTimeMutex: &sync.Mutex{},
		stopChan:            make(chan struct{}),
	}

	// Inicia a goroutine para fechamento automático
	go repo.startAutoCloseRoutine()

	return repo
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	// Armazena o tempo de fim do leilão para controle
	ar.auctionEndTimeMutex.Lock()
	ar.auctionEndTimeMap[auctionEntity.Id] = auctionEntity.Timestamp.Add(ar.auctionInterval)
	ar.auctionEndTimeMutex.Unlock()

	return nil
}

// UpdateAuctionStatus atualiza o status de um leilão
func (ar *AuctionRepository) UpdateAuctionStatus(
	ctx context.Context,
	auctionId string,
	status auction_entity.AuctionStatus) *internal_error.InternalError {
	filter := bson.M{"_id": auctionId}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error trying to update auction status", err)
		return internal_error.NewInternalServerError("Error trying to update auction status")
	}

	return nil
}

// getAuctionInterval obtém o intervalo de tempo do leilão das variáveis de ambiente
func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5 // Default de 5 minutos
	}
	return duration
}

// getCheckInterval obtém o intervalo de verificação de leilões expirados das variáveis de ambiente
func getCheckInterval() time.Duration {
	checkInterval := os.Getenv("AUCTION_CHECK_INTERVAL")
	duration, err := time.ParseDuration(checkInterval)
	if err != nil {
		return time.Second * 10 // Default de 10 segundos
	}
	return duration
}

// getContextTimeout obtém o timeout para operações de contexto das variáveis de ambiente
func getContextTimeout() time.Duration {
	contextTimeout := os.Getenv("AUCTION_CONTEXT_TIMEOUT")
	duration, err := time.ParseDuration(contextTimeout)
	if err != nil {
		return time.Second * 30 // Default de 30 segundos
	}
	return duration
}

// startAutoCloseRoutine inicia a goroutine que verifica e fecha leilões automaticamente
func (ar *AuctionRepository) startAutoCloseRoutine() {
	ticker := time.NewTicker(ar.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ar.checkAndCloseExpiredAuctions()
		case <-ar.stopChan:
			return
		}
	}
}

// checkAndCloseExpiredAuctions verifica e fecha leilões expirados
func (ar *AuctionRepository) checkAndCloseExpiredAuctions() {
	ctx, cancel := context.WithTimeout(context.Background(), ar.contextTimeout)
	defer cancel()

	now := time.Now()

	ar.auctionEndTimeMutex.Lock()
	expiredAuctions := make([]string, 0)

	for auctionId, endTime := range ar.auctionEndTimeMap {
		if now.After(endTime) {
			expiredAuctions = append(expiredAuctions, auctionId)
		}
	}
	ar.auctionEndTimeMutex.Unlock()

	// Fecha os leilões expirados
	for _, auctionId := range expiredAuctions {
		err := ar.UpdateAuctionStatus(ctx, auctionId, auction_entity.Completed)
		if err != nil {
			logger.Error("Error closing expired auction", err)
		} else {
			logger.Info("Auction closed automatically",
				zap.String("auctionId", auctionId),
				zap.Time("timestamp", now),
			)

			// Remove do mapa de controle
			ar.auctionEndTimeMutex.Lock()
			delete(ar.auctionEndTimeMap, auctionId)
			ar.auctionEndTimeMutex.Unlock()
		}
	}
}

// StopAutoCloseRoutine para a goroutine de fechamento automático
func (ar *AuctionRepository) StopAutoCloseRoutine() {
	close(ar.stopChan)
}
