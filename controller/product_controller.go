package controller

import (
	"api/models"
	"api/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type productController struct {
	productUseCase usecase.ProductUsecase
}

func NewProductController(usecase usecase.ProductUsecase) productController {
	return productController{
		productUseCase: usecase,
	}
}

func (p *productController) GetProducts(ctx *gin.Context) {

	products, err := p.productUseCase.GetProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, products)
}

func (p *productController) CreateProduct(ctx *gin.Context) {

	var product models.Product
	err := ctx.BindJSON(&product)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	insertedProduct, err := p.productUseCase.CreateProduct(product)
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
