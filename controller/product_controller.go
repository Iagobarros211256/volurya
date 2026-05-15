package controller

import (
	"api/jobs"
	"api/logger"
	"api/models"
	"api/repository"
	"api/storage"
	"api/usecase"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productUsecase *usecase.ProductUsecase
	productRepo    *repository.ProductRepository
	r2Storage      *storage.R2Storage
	imageProcessor *jobs.ImageProcessor // NOVO
}

// No NewProductController
func NewProductController(
	productUsecase *usecase.ProductUsecase,
	productRepo *repository.ProductRepository,
	r2Storage *storage.R2Storage,
	imageProcessor *jobs.ImageProcessor,
) *ProductController { // MAIÚSCULO
	return &ProductController{ // MAIÚSCULO
		productUsecase: productUsecase,
		productRepo:    productRepo,
		r2Storage:      r2Storage,
		imageProcessor: imageProcessor,
	}
}

func (p *ProductController) GetProducts(ctx *gin.Context) {

	limit := ctx.Query("limit")
	cursor := ctx.Query("cursor")

	products, nextCursor, hasMore, err :=
		p.productUsecase.GetProducts(limit, cursor)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": products,
		"pagination": gin.H{
			"next_cursor": nextCursor,
			"has_more":    hasMore,
		},
	})
}

func (p *ProductController) CreateProduct(ctx *gin.Context) {

	var product models.Product
	if err := ctx.BindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	userID := ctx.GetInt("user_id") //  vem do middleware JWT

	insertedProduct, err :=
		p.productUsecase.CreateProduct(userID, product)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, insertedProduct)
}

func (p *ProductController) GetProductById(ctx *gin.Context) {

	id := ctx.Param("productId")
	if id == "" {
		response := models.Response{
			Message: "Id do produto nao pode ser nulo",
		}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	productId, err := strconv.Atoi(id)
	if err != nil {
		response := models.Response{
			Message: "Id do produto precisa ser um numero",
		}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	product, err := p.productUsecase.GetProductById(productId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if product == nil {
		response := models.Response{
			Message: "Produto nao foi encontrado na base de dados",
		}
		ctx.JSON(http.StatusNotFound, response)
		return
	}

	ctx.JSON(http.StatusOK, product)
}

func (p *ProductController) UpdateProduct(ctx *gin.Context) {

	id := ctx.Param("productId")
	if id == "" {
		response := models.Response{
			Message: "Id do produto nao pode ser nulo",
		}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	productId, err := strconv.Atoi(id)
	if err != nil {
		response := models.Response{
			Message: "Id do produto precisa ser um numero",
		}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Stock       int     `json:"stock"`
	}

	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, models.Response{
			Message: "Body invalido",
		})
		return
	}

	product, err := p.productUsecase.UpdateProduct(productId, body.Name, body.Description, body.Price, body.Stock)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	if product == nil {
		response := models.Response{
			Message: "Produto nao foi encontrado na base de dados",
		}
		ctx.JSON(http.StatusNotFound, response)
		return
	}

	ctx.JSON(http.StatusOK, product)
}

func (p *ProductController) Delete(ctx *gin.Context) {

	idParam := ctx.Param("productId")
	if idParam == "" {
		ctx.JSON(http.StatusBadRequest, models.Response{
			Message: "Id do produto nao pode ser nulo",
		})
		return
	}

	productId, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Response{
			Message: "Id do produto precisa ser um numero",
		})
		return
	}

	userID := ctx.GetInt("user_id") //  JWT

	err = p.productUsecase.Delete(userID, productId)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, models.Response{
				Message: "Produto nao encontrado",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, models.Response{
			Message: err.Error(),
		})
		return
	}

	ctx.Status(http.StatusNoContent) // 204
}

func (pc *ProductController) UploadImage(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil {
		c.JSON(400, gin.H{"error": "ID inválido"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "arquivo não encontrado"})
		return
	}

	// Validar tipo
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		// "image/webp": true,  // Remover por enquanto
	}

	if !allowedTypes[file.Header.Get("Content-Type")] {
		c.JSON(400, gin.H{"error": "tipo de arquivo não permitido"})
		return
	}

	// Validar tamanho
	const maxSize = 5 * 1024 * 1024 // 5MB
	if file.Size > maxSize {
		c.JSON(400, gin.H{"error": "arquivo muito grande (máximo 5MB)"})
		return
	}

	// Ler dados do arquivo
	fileData, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "erro ao ler arquivo"})
		return
	}
	defer fileData.Close()

	buf := &bytes.Buffer{}
	io.Copy(buf, fileData)

	// Enfileirar para processamento assíncrono
	job := jobs.ImageJob{
		ProductID:  productID,
		FileName:   file.Filename,
		FileData:   buf.Bytes(),
		TargetSize: maxSize,
		OnComplete: func(processedData []byte, err error) {
			if err != nil {
				logger.Log.LogAttrs(
					c.Request.Context(),
					slog.LevelError,
					"Image processing failed",
					slog.Int("product_id", productID),
					slog.String("error", err.Error()),
				)
				return
			}

			// Após processamento: fazer upload para R2
			ext := "jpg" // sempre salvamos como JPEG após processamento
			key := fmt.Sprintf("products/%d.%s", time.Now().UnixNano(), ext)
			publicURL, uploadErr := pc.r2Storage.UploadImageBytes(key, processedData)
			if uploadErr != nil {
				logger.Log.LogAttrs(
					c.Request.Context(),
					slog.LevelError,
					"R2 upload failed",
					slog.Int("product_id", productID),
					slog.String("error", uploadErr.Error()),
				)
				return
			}

			// Atualizar produto no banco
			product, updateErr := pc.productRepo.UpdateImageURL(productID, publicURL)
			if updateErr != nil {
				logger.Log.LogAttrs(
					c.Request.Context(),
					slog.LevelError,
					"Failed to update product image URL",
					slog.Int("product_id", productID),
					slog.String("error", updateErr.Error()),
				)
				return
			}

			logger.Log.LogAttrs(
				c.Request.Context(),
				slog.LevelInfo,
				"Image uploaded successfully",
				slog.Int("product_id", productID),
				slog.String("url", product.ImageURL),
			)
		},
	}

	err = pc.imageProcessor.ProcessAsync(job)
	if err != nil {
		c.JSON(500, gin.H{"error": "fila de processamento cheia, tente novamente"})
		return
	}

	// Retornar imediatamente sem esperar o processamento
	c.JSON(202, gin.H{
		"message":    "Imagem enviada para processamento",
		"product_id": productID,
	})
}
