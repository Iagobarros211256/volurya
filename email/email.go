package email

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client  *resend.Client
	from    string
	enabled bool
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		slog.Warn("RESEND_API_KEY not set — email notifications disabled")
		return &EmailService{enabled: false}
	}

	return &EmailService{
		client:  resend.NewClient(apiKey),
		from:    "Volurya Shop <noreply@volurya.com>",
		enabled: true,
	}
}

// SendOrderConfirmation envia email de confirmação de pedido ao cliente.
func (s *EmailService) SendOrderConfirmation(toEmail string, orderID int, amount int, items []OrderItem) error {
	if !s.enabled {
		slog.Info("email service disabled, skipping order confirmation",
			"order_id", orderID,
			"to", toEmail,
		)
		return nil
	}

	html := buildOrderConfirmationHTML(orderID, amount, items)

	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Pedido #%d confirmado — Volurya Shop", orderID),
		Html:    html,
	}

	resp, err := s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send order confirmation email: %w", err)
	}

	slog.Info("order confirmation email sent",
		"order_id", orderID,
		"to", toEmail,
		"email_id", resp.Id,
	)

	return nil
}

// OrderItem representa um item para o email de confirmação.
type OrderItem struct {
	Name      string
	Quantity  int
	UnitPrice float64
}

func buildOrderConfirmationHTML(orderID int, amountCents int, items []OrderItem) string {
	total := float64(amountCents) / 100

	itemsHTML := ""
	for _, item := range items {
		itemsHTML += fmt.Sprintf(`
			<tr>
				<td style="padding: 8px; border-bottom: 1px solid #333;">%s</td>
				<td style="padding: 8px; border-bottom: 1px solid #333; text-align: center;">%d</td>
				<td style="padding: 8px; border-bottom: 1px solid #333; text-align: right;">R$ %.2f</td>
			</tr>`,
			item.Name, item.Quantity, item.UnitPrice*float64(item.Quantity),
		)
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="background-color: #0a0a0a; color: #ffffff; font-family: Arial, sans-serif; margin: 0; padding: 20px;">
  <div style="max-width: 600px; margin: 0 auto;">

    <!-- Header -->
    <div style="text-align: center; padding: 40px 0 20px;">
      <h1 style="color: #dc3545; font-size: 2rem; margin: 0; letter-spacing: 4px;">VOLURYA</h1>
      <p style="color: #666; margin: 5px 0 0;">Official Merchandise</p>
    </div>

    <!-- Confirmação -->
    <div style="background: #1a1a1a; border: 1px solid #dc3545; border-radius: 8px; padding: 30px; margin: 20px 0;">
      <h2 style="color: #dc3545; margin: 0 0 10px;">Pedido Confirmado!</h2>
      <p style="color: #ccc; margin: 0;">Pedido <strong style="color: #fff;">#%d</strong> — obrigado pela compra!</p>
    </div>

    <!-- Itens -->
    <div style="background: #1a1a1a; border-radius: 8px; padding: 20px; margin: 20px 0;">
      <h3 style="color: #fff; margin: 0 0 15px;">Itens do pedido</h3>
      <table style="width: 100%%; border-collapse: collapse;">
        <thead>
          <tr>
            <th style="text-align: left; padding: 8px; border-bottom: 1px solid #dc3545; color: #dc3545;">Produto</th>
            <th style="text-align: center; padding: 8px; border-bottom: 1px solid #dc3545; color: #dc3545;">Qtd</th>
            <th style="text-align: right; padding: 8px; border-bottom: 1px solid #dc3545; color: #dc3545;">Total</th>
          </tr>
        </thead>
        <tbody>
          %s
        </tbody>
        <tfoot>
          <tr>
            <td colspan="2" style="padding: 12px 8px 0; color: #fff; font-weight: bold;">Total</td>
            <td style="padding: 12px 8px 0; text-align: right; color: #dc3545; font-weight: bold; font-size: 1.1rem;">R$ %.2f</td>
          </tr>
        </tfoot>
      </table>
    </div>

    <!-- Footer -->
    <div style="text-align: center; padding: 20px 0; color: #666; font-size: 0.85rem;">
      <p>Volurya Shop — Fortaleza, Brasil</p>
      <p>Dúvidas? Entre em contato: <a href="mailto:contato@volurya.com" style="color: #dc3545;">contato@volurya.com</a></p>
    </div>

  </div>
</body>
</html>`, orderID, itemsHTML, total)
}
