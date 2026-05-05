package controller

import (
	"api/models"
	"api/storage"
	"api/usecase"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type productController struct {
	productUseCase *usecase.ProductUsecase
	r2             *storage.R2Storage
}

func NewProductController(usecase *usecase.ProductUsecase, r2 *storage.R2Storage) *productController {
	return &productController{
		productUseCase: usecase,
		r2:             r2,
	}
}

func (p *productController) GetProducts(ctx *gin.Context) {

	limit := ctx.Query("limit")
	cursor := ctx.Query("cursor")

	products, nextCursor, hasMore, err :=
		p.productUseCase.GetProducts(limit, cursor)

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

func (p *productController) CreateProduct(ctx *gin.Context) {

	var product models.Product
	if err := ctx.BindJSON(&product); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	userID := ctx.GetInt("user_id") //  vem do middleware JWT

	insertedProduct, err :=
		p.productUseCase.CreateProduct(userID, product)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, insertedProduct)
}

func (p *productController) GetProductById(ctx *gin.Context) {

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

	product, err := p.productUseCase.GetProductById(productId)
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

func (p *productController) UpdateProduct(ctx *gin.Context) {

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

	product, err := p.productUseCase.UpdateProduct(productId, body.Name, body.Description, body.Price, body.Stock)
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

func (p *productController) Delete(ctx *gin.Context) {

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

	err = p.productUseCase.Delete(userID, productId)
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

func (p *productController) UploadImage(ctx *gin.Context) {
	userID := ctx.GetInt("user_id")

	productId, err := strconv.Atoi(ctx.Param("productId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Response{Message: "Id do produto precisa ser um numero"})
		return
	}

	file, header, err := ctx.Request.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}
	defer file.Close()

	product, err := p.productUseCase.UploadImage(userID, productId, file, header, p.r2)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, product)
}
