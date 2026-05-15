package usecase

import (
	"api/jobs"
	"api/logger"
	"api/metrics"
	"api/models"
	"api/repository"
	"api/storage"
	"context"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"strconv"
)

func NewProductUsecase(repo *repository.ProductRepository) *ProductUsecase {
	return &ProductUsecase{
		repo: repo,
	}
}

// NewProductUsecaseWithImageProcessor cria um nova instância com suporte a processamento de imagens
func NewProductUsecaseWithImageProcessor(repo *repository.ProductRepository,
	imageProcessor *jobs.ImageProcessor, storageProvider *storage.R2Storage) *ProductUsecase {
	return &ProductUsecase{
		repo:            repo,
		imageProcessor:  imageProcessor,
		storageProvider: storageProvider,
	}
}

type ProductUsecase struct {
	repo            *repository.ProductRepository
	imageProcessor  *jobs.ImageProcessor
	storageProvider *storage.R2Storage
}

func (pu *ProductUsecase) GetProducts(limitStr string, cursorStr string) ([]models.Product, *int, bool, error) {

	const (
		defaultLimit = 10
		maxLimit     = 50
	)

	limit := defaultLimit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			return nil, nil, false, errors.New("invalid limit")
		}
		if parsedLimit > maxLimit {
			limit = maxLimit
		} else {
			limit = parsedLimit
		}
	}

	var cursor *int
	if cursorStr != "" {
		parsedCursor, err := strconv.Atoi(cursorStr)
		if err != nil || parsedCursor < 0 {
			return nil, nil, false, errors.New("invalid cursor")
		}
		cursor = &parsedCursor
	}

	products, hasMore, err :=
		pu.repo.GetProducts(limit, cursor) // ← mudou de repository para repo
	if err != nil {
		return nil, nil, false, err
	}

	var nextCursor *int
	if hasMore && len(products) > 0 {
		lastID := products[len(products)-1].ID
		nextCursor = &lastID
	}

	return products, nextCursor, hasMore, nil
}

func (pu *ProductUsecase) CreateProduct(userID int, product models.Product) (models.Product, error) {

	productId, err := pu.repo.CreateProduct(product, userID) // ← mudou para repo
	if err != nil {
		return models.Product{}, err
	}

	product.ID = productId
	product.UserID = userID
	return product, nil
}

func (pu *ProductUsecase) GetProductById(id_product int) (*models.Product, error) {

	product, err := pu.repo.GetProductById(id_product) // ← mudou para repo
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (pu *ProductUsecase) UpdateProduct(id_product int, Name string, Description string, Price float64, Stock int) (*models.Product, error) {
	// essas regras nao serao necessarias agora mas no futuro e num ambiente de producao serao indispensaveis
	if Price < 0 {
		return nil, errors.New("price cannot be negative")
	}

	if Stock < 0 {
		return nil, errors.New("stock cannot be negative")
	}

	product, err := pu.repo.UpdateProduct(id_product, Name, Description, Price, Stock) // ← mudou para repo
	if err != nil {
		return nil, err
	}

	return product, nil
}

// delete one
func (pu *ProductUsecase) Delete(userID int, productID int) error {
	product, err := pu.repo.GetProductById(productID)
	if err != nil {
		return err
	}

	if product == nil {
		return errors.New("product not found")
	}

	// Checagem de ownership
	if product.UserID != userID {
		return errors.New("forbidden: you do not own this product")
	}

	// Regra futura para admin (opcional)
	// if userRole == "admin" { ok }

	return pu.repo.Delete(productID)
}

func (pu *ProductUsecase) UploadImage(userID, productID int, file multipart.File, header *multipart.FileHeader, store *storage.R2Storage) (*models.Product, error) {
	// Verifica se o produto existe e pertence ao usuário
	product, err := pu.repo.GetProductById(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if product.UserID != userID {
		return nil, errors.New("forbidden: you do not own this product")
	}

	// Se imageProcessor não está configurado, usa upload síncrono (fallback)
	if pu.imageProcessor == nil {
		imageURL, err := store.UploadImage(file, header)
		if err != nil {
			return nil, err
		}

		metrics.ImageUploadsTotal.Inc()

		updated, err := pu.repo.UpdateImageURL(productID, imageURL)
		if err != nil {
			return nil, err
		}

		return updated, nil
	}

	// ===== ASYNC PIPELINE PROCESSING =====
	// 1. Ler arquivo para memória
	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(fileData) == 0 {
		return nil, errors.New("arquivo vazio")
	}

	// 2. Criar callback para atualizar banco após processamento
	onImageComplete := pu.createImageProcessingCallback(productID, userID)

	// 3. Criar callback para upload em R2 e atualização do banco
	storageCallback := func(processedData []byte) (string, error) {
		publicURL, err := store.UploadProcessedImage(productID, processedData)
		if err != nil {
			logger.Log.LogAttrs(
				context.Background(),
				slog.LevelError,
				"Failed to upload processed image to R2",
				slog.Int("product_id", productID),
				slog.String("error", err.Error()),
			)
			return "", err
		}

		// Atualizar banco com nova URL
		_, err = pu.repo.UpdateImageURL(productID, publicURL)
		if err != nil {
			logger.Log.LogAttrs(
				context.Background(),
				slog.LevelError,
				"Failed to update product image URL in database",
				slog.Int("product_id", productID),
				slog.String("url", publicURL),
				slog.String("error", err.Error()),
			)
			return publicURL, err // Retorna URL mesmo com erro do DB (pode ser retentado)
		}

		logger.Log.LogAttrs(
			context.Background(),
			slog.LevelInfo,
			"Image successfully processed and stored",
			slog.Int("product_id", productID),
			slog.String("url", publicURL),
		)

		return publicURL, nil
	}

	// 4. Enfileirar job de processamento assíncrono
	imageJob := jobs.ImageJob{
		ProductID:          productID,
		FileName:           header.Filename,
		FileData:           fileData,
		TargetSize:         5 * 1024 * 1024, // 5MB max
		ResizeWidth:        1200,            // Redimensiona para 1200x1200 max
		ResizeHeight:       1200,
		CompressionQuality: 85,
		OnComplete:         onImageComplete,
		StorageCallback:    storageCallback,
	}

	err = pu.imageProcessor.ProcessAsync(imageJob)
	if err != nil {
		return nil, err
	}

	metrics.ImageUploadsTotal.Inc()

	// Retorna o produto atual (processamento acontece assincronamente)
	// URL será atualizada no banco quando o processamento terminar
	return product, nil
}

// createImageProcessingCallback cria um callback para ser executado após o processamento
func (pu *ProductUsecase) createImageProcessingCallback(productID, userID int) func([]byte, error) {
	return func(processedData []byte, err error) {
		ctx := context.Background()

		if err != nil {
			logger.Log.LogAttrs(
				ctx,
				slog.LevelError,
				"Image processing failed - will not update database",
				slog.Int("product_id", productID),
				slog.String("error", err.Error()),
			)
			return
		}

		logger.Log.LogAttrs(
			ctx,
			slog.LevelInfo,
			"Image processing completed successfully",
			slog.Int("product_id", productID),
			slog.Int("processed_size", len(processedData)),
		)

		// Note: URL é atualizada pelo StorageCallback durante o processamento
		// Este callback é apenas para logging de conclusão
	}
}
