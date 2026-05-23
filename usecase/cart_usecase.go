package usecase

import (
	"api/metrics"
	"api/models"
	"api/repository"
	"errors"
)

type CartUsecase struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartUsecase(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartUsecase {
	return &CartUsecase{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// GetCart retorna o carrinho do usuário com todos os itens
func (uc *CartUsecase) GetCart(userID int) (*models.Cart, error) {
	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	items, err := uc.cartRepo.GetCartItems(cart.ID)
	if err != nil {
		return nil, err
	}

	cart.Items = items
	return cart, nil
}

// AddItem adiciona um produto ao carrinho
func (uc *CartUsecase) AddItem(userID, productID, quantity int) (*models.CartItem, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// Verifica se o produto existe e tem estoque
	product, err := uc.productRepo.GetProductById(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if product.Stock < quantity {
		return nil, ErrInsufficientStock
	}

	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	item, err := uc.cartRepo.AddItem(cart.ID, productID, quantity)
	if err != nil {
		return nil, err
	}

	metrics.CartItemsAdded.Inc()

	item.Product = product
	return item, nil
}

// UpdateItem atualiza a quantidade de um item do carrinho
func (uc *CartUsecase) UpdateItem(userID, itemID, quantity int) (*models.CartItem, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// Verifica se o item pertence ao carrinho do usuário
	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	item, err := uc.cartRepo.GetItemByID(itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("item not found")
	}
	if item.CartID != cart.ID {
		return nil, errors.New("forbidden: item does not belong to your cart")
	}

	// Verifica estoque
	product, err := uc.productRepo.GetProductById(item.ProductID)
	if err != nil {
		return nil, err
	}
	if product.Stock < quantity {
		return nil, ErrInsufficientStock
	}

	updated, err := uc.cartRepo.UpdateItemQuantity(itemID, quantity)
	if err != nil {
		return nil, err
	}

	updated.Product = product
	return updated, nil
}

// RemoveItem remove um item do carrinho
func (uc *CartUsecase) RemoveItem(userID, itemID int) error {
	cart, err := uc.cartRepo.GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	item, err := uc.cartRepo.GetItemByID(itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("item not found")
	}
	if item.CartID != cart.ID {
		return errors.New("forbidden: item does not belong to your cart")
	}

	return uc.cartRepo.RemoveItem(itemID)
}

// Checkout finaliza o carrinho e cria uma ordem por produto
func (uc *CartUsecase) Checkout(userID int, orderUsecase *OrderUsecase) ([]string, error) {
	cart, err := uc.GetCart(userID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 {
		return nil, errors.New("cart is empty")
	}

	paymentURLs := make([]string, 0)
	for _, item := range cart.Items {
		url, err := orderUsecase.CreateOrder(userID, item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		paymentURLs = append(paymentURLs, url)
	}

	// Limpa o carrinho após checkout
	if err := uc.cartRepo.ClearCart(cart.ID); err != nil {
		return nil, err
	}

	return paymentURLs, nil
}

/*

Dependências concretas
gocartRepo    *repository.CartRepository
productRepo *repository.ProductRepository
Padrão recorrente — sem interfaces, sem testes unitários possíveis.

🔴 Checkout recebe *OrderUsecase como parâmetro
gofunc (uc *CartUsecase) Checkout(userID int, orderUsecase *OrderUsecase) ([]string, error) {
Já apontado no cart_controller.go — acoplamento circular entre usecases. CartUsecase dependendo de OrderUsecase como parâmetro de método é pior do que injeção via construtor. Injete via construtor ou crie um CheckoutUsecase dedicado.

🔴 Checkout sem transação — estado inconsistente possível
gofor _, item := range cart.Items {
    url, err := orderUsecase.CreateOrder(...)
    if err != nil {
        return nil, err  // ordens anteriores já foram criadas
    }
}
// Limpa o carrinho após checkout
uc.cartRepo.ClearCart(cart.ID)
Se CreateOrder falhar no terceiro item de quatro, dois pedidos foram criados mas o carrinho não foi limpo. Se ClearCart falhar após todas as ordens criadas, o usuário tem ordens duplicadas no próximo checkout. Envolva em transação ou implemente compensação.

🔴 AddItem não verifica estoque total do carrinho
goif product.Stock < quantity {
    return nil, ErrInsufficientStock
}
Verifica apenas se o estoque é suficiente para a quantidade adicionada agora, mas não considera itens já no carrinho. Se há 5 em estoque e o usuário adiciona 3, depois mais 3, passa na validação mas o checkout vai falhar. Verifique a quantidade total:
go// Verificar quantidade já no carrinho + nova quantidade
existingItem, _ := uc.cartRepo.GetItemByProductID(cart.ID, productID)
totalQuantity := quantity
if existingItem != nil {
    totalQuantity += existingItem.Quantity
}
if product.Stock < totalQuantity {
    return nil, ErrInsufficientStock
}

🔴 product == nil após GetProductById com erro sentinela
goproduct, err := uc.productRepo.GetProductById(productID)
if err != nil {
    return nil, err
}
if product == nil {
    return nil, errors.New("product not found")
}
ProductRepository.GetProductById retorna ErrProductNotFound quando não encontrado — nunca retorna nil, nil. A verificação product == nil nunca vai ser verdadeira. Use:
goif errors.Is(err, repository.ErrProductNotFound) {
    return nil, errors.New("product not found")
}

🟡 GetOrCreateCart chamado desnecessariamente em RemoveItem e UpdateItem
gocart, err := uc.cartRepo.GetOrCreateCart(userID)
Se o usuário não tem carrinho, GetOrCreateCart cria um — mas um carrinho vazio nunca vai ter o item sendo removido/atualizado. Use GetCart sem criar:
go// No repository: GetCartByUser sem criar
cart, err := uc.cartRepo.GetCartByUser(userID)
if err != nil || cart == nil {
    return nil, errors.New("cart not found")
}

🟡 Erros sem wrap e sem erros sentinela
goreturn nil, errors.New("item not found")
return nil, errors.New("cart is empty")
return nil, errors.New("forbidden: item does not belong to your cart")
Strings de erro sem sentinelas dificultam tratamento no controller. Adicione em errors.go:
govar ErrItemNotFound = errors.New("item not found")
var ErrCartEmpty = errors.New("cart is empty")
var ErrItemNotInCart = errors.New("item does not belong to cart")

🟡 Checkout cria uma ordem por produto em vez de uma ordem com múltiplos itens
gofor _, item := range cart.Items {
    url, err := orderUsecase.CreateOrder(userID, item.ProductID, item.Quantity)
Isso gera N ordens para N produtos no carrinho — o usuário recebe múltiplos links de pagamento. O design correto é uma única ordem com múltiplos order_items, gerando um único checkout.

*/
