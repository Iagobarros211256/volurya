package models

import "time"

type Cart struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	CreatedAt time.Time  `json:"created_at"`
	Items     []CartItem `json:"items"`
}

type CartItem struct {
	ID        int       `json:"id"`
	CartID    int       `json:"cart_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	AddedAt   time.Time `json:"added_at"`
	Product   *Product  `json:"product,omitempty"`
}

/*

Sem UpdatedAt no Cart
Já apontado na migration — sem updated_at não há como rastrear carrinhos abandonados ou última atividade:
gotype Cart struct {
    ID        int        `json:"id"`
    UserID    int        `json:"user_id"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    Items     []CartItem `json:"items"`
}

🟡 AddedAt inconsistente com o resto do projeto
Todas as outras models usam CreatedAt. CartItem usa AddedAt — mesma inconsistência apontada na migration:
goCreatedAt time.Time `json:"created_at"`

🟡 Sem campo de preço total no Cart
O carrinho não tem TotalPrice calculado. O frontend provavelmente precisa recalcular somando item.Product.Price * item.Quantity para cada item — lógica de negócio vazando para o cliente:
gotype Cart struct {
    // ...
    TotalPrice float64    `json:"total_price"`
    ItemCount  int        `json:"item_count"`
}

🟢 CartID no CartItem é redundante na resposta
Quando CartItem é retornado dentro de Cart.Items, o CartID já é implícito. Considere omitir na serialização:
goCartID int `json:"cart_id,omitempty"`

*/
