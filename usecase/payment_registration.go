package usecase

import "fmt"

// RegisterPaymentIntentAtomically registra o vínculo entre pedido
// e PaymentIntent junto com o registro financeiro, de forma atômica.
func (pu *PaymentUsecase) RegisterPaymentIntentAtomically(
	orderID int,
	paymentIntentID string,
	amount int,
	currency string,
	userEmail string,
) error {
	err := pu.orderRepo.RegisterPaymentIntentAtomically(
		orderID,
		paymentIntentID,
		amount,
		currency,
		userEmail,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to register payment intent atomically: %w",
			err,
		)
	}

	return nil
}
