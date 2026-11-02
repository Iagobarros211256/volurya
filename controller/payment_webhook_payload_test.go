package controller

import "testing"

func TestExtractPaymentIntentIDFromStripeRawObject(t *testing.T) {
	raw := []byte(`{
		"id": "pi_test_123",
		"object": "payment_intent",
		"status": "succeeded"
	}`)

	got, err := extractPaymentIntentID(raw)
	if err != nil {
		t.Fatalf("falha ao extrair PaymentIntent válido: %v", err)
	}

	if got != "pi_test_123" {
		t.Errorf(
			"PaymentIntent incorreto: recebido %q, esperado %q",
			got, "pi_test_123",
		)
	}
}

func TestExtractPaymentErrorMessageFromStripeRawObject(t *testing.T) {
	raw := []byte(`{
		"id": "pi_test_declined",
		"object": "payment_intent",
		"status": "requires_payment_method",
		"last_payment_error": {
			"type": "card_error",
			"code": "card_declined",
			"message": "Your card was declined."
		}
	}`)

	got := extractPaymentErrorMessage(raw)
	want := "Your card was declined."

	if got != want {
		t.Errorf(
			"mensagem de erro incorreta: recebido %q, esperado %q",
			got, want,
		)
	}
}
