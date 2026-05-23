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

/*

Dependências concretas
gorepo            *repository.ProductRepository
imageProcessor  *jobs.ImageProcessor
storageProvider *storage.R2Storage
Padrão recorrente — sem interfaces.

🔴 Dois construtores com responsabilidades diferentes
gofunc NewProductUsecase(repo *repository.ProductRepository) *ProductUsecase
func NewProductUsecaseWithImageProcessor(...) *ProductUsecase
O main.go usa NewProductUsecase mas passa imageProcessor separadamente para o controller — o processador nunca é injetado no usecase via construtor. NewProductUsecaseWithImageProcessor existe mas não é usado. Unifique:
gofunc NewProductUsecase(repo *repository.ProductRepository, imageProcessor *jobs.ImageProcessor, storage *storage.R2Storage) *ProductUsecase

🔴 GetProducts recebe string e converte para int
gofunc (pu *ProductUsecase) GetProducts(limitStr string, cursorStr string) (...)
Já apontado no product_controller.go — o controller converte string→int, passa para o usecase que converte de volta. O usecase deveria receber int e *int:
gofunc (pu *ProductUsecase) GetProducts(limit int, cursor *int) ([]models.Product, *int, bool, error)

🔴 Delete verifica ownership via product.UserID
goif product.UserID != userID {
    return errors.New("forbidden: you do not own this product")
}
Como apontado na migration, user_id em produtos é design questionável. Além disso, admin deveria poder deletar qualquer produto — o comentário sugere isso mas está desabilitado:
go// Regra futura para admin (opcional)
// if userRole == "admin" { ok }
O middleware RequireAdminRole já garante que só admins chegam aqui — a verificação de ownership é redundante e bloqueia admins de deletar produtos de outros usuários.

🔴 storageCallback retorna URL mesmo com erro de banco
goreturn publicURL, err // Retorna URL mesmo com erro do DB (pode ser retentado)
Imagem foi upada no R2 mas o banco não foi atualizado — estado inconsistente. A URL existe no storage mas o produto não a referencia. Sem mecanismo de retry isso fica perdido:
go// Sem retry implementado, isso cria dados órfãos no R2

🟡 OnComplete e StorageCallback fazem coisas sobrepostas
Já apontado no image_worker.go — dois callbacks para o mesmo fluxo. StorageCallback já faz o upload e atualiza o banco. OnComplete só loga. O design está confuso — simplifique para um único callback.

🟡 UploadImage recebe store *storage.R2Storage como parâmetro
gofunc (pu *ProductUsecase) UploadImage(..., store *storage.R2Storage) (*models.Product, error)
storageProvider já está no struct mas store é passado como parâmetro — qual é usado? No path síncrono usa store, no async usa store também. pu.storageProvider nunca é usado:
go// storageProvider no struct é inútil — sempre usa o parâmetro

🟡 product == nil após erro sentinela
goproduct, err := pu.repo.GetProductById(productID)
if err != nil {
    return nil, err
}
if product == nil {  // nunca verdadeiro com ErrProductNotFound
    return nil, errors.New("product not found")
}
Mesmo problema do cart_usecase.go — GetProductById retorna ErrProductNotFound, nunca nil, nil.

🟡 Nomenclatura inconsistente — id_product, Name, Description
gofunc (pu *ProductUsecase) UpdateProduct(id_product int, Name string, Description string, ...)
func (pu *ProductUsecase) GetProductById(id_product int)
Go usa camelCase. id_product deveria ser productID, Name deveria ser name, Description deveria ser description.

🟡 Comentários de desenvolvimento no código
goproducts, hasMore, err := pu.repo.GetProducts(limit, cursor) // ← mudou de repository para repo
productId, err := pu.repo.CreateProduct(product, userID) // ← mudou para repo
// essas regras nao serao necessarias agora mas no futuro...

*/
