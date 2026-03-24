package usecase

import (
  "api/models"
  "api/repository"
  "github.com/WalterPaes/go-pagseguro"
  "os"
)

type OrderUsecase struct {
  repo *repository.OrderRepository
}

func NewOrderUsecase(repo *repository.OrderRepository) *OrderUsecase {
  return &OrderUsecase{repo: repo}
}

func (ou *OrderUsecase) CreateOrder(userID int, productID int, quantity int) (string, error) {
  // Pegue produto
  product, err := ou.repo.GetProductById(productID)
  if err != nil {
    return "", err
  }

  // Crie ordem no banco
  order := models.Order{
    UserID: userID,
    ProductID: productID,
    Quantity: quantity,
    Total: product.Price * float64(quantity),
    Status: "pending",
  }
  orderID, err := ou.repo.CreateOrder(order)
  if err != nil {
    return "", err
  }

  // Integre com PagSeguro
  config := pagseguro.Config{
    Url: "https://sandbox.api.pagseguro.com" if os.Getenv("PAGSEGURO_SANDBOX") == "true" else "https://api.pagseguro.com",
    Token: os.Getenv("PAGSEGURO_TOKEN"),
    Email: os.Getenv("PAGSEGURO_EMAIL"),
  }
  client := pagseguro.NewClient(config)

  charge := pagseguro.Charge{
    ReferenceID: fmt.Sprintf("ORDER-%d", orderID),
    Description: "Compra Volurya",
    Amount: pagseguro.Amount{
      Value: int(order.Total * 100),  // em centavos
      Currency: "BRL",
    },
    PaymentMethod: pagseguro.PaymentMethod{
      Type: "CREDIT_CARD",  // ou PIX, BOLETO
    },
    NotificationUrls: []string{os.Getenv("PAGSEGURO_WEBHOOK_URL")},
  }

  chargeRes, err := client.Charges.Create(charge)
  if err != nil {
    return "", err
  }

  // Atualize ordem no banco com charge ID
  ou.repo.UpdateOrderChargeID(orderID, chargeRes.ID)

  return chargeRes.PaymentLinks[0].Href, nil  // link pra pagar
}