package mailer

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type Mailer struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Enabled  bool
}

func NewMailer() *Mailer {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := os.Getenv("MAIL_FROM")

	if from == "" {
		from = "nao-responda@imovxen.com.br"
	}

	port := 587
	if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
		port = p
	}

	enabled := host != "" && user != ""

	return &Mailer{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		From:     from,
		Enabled:  enabled,
	}
}

// Send envia um e-mail com conteúdo HTML de forma assíncrona
func (m *Mailer) Send(to string, subject string, htmlBody string) {
	go func() {
		if !m.Enabled {
			slog.Info("[EMAIL SIMULADO LOCAL]", "para", to, "assunto", subject)
			return
		}

		headers := make(map[string]string)
		headers["From"] = fmt.Sprintf("Imovxen <%s>", m.From)
		headers["To"] = to
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"

		var message strings.Builder
		for k, v := range headers {
			message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		message.WriteString("\r\n" + htmlBody)

		auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
		addr := fmt.Sprintf("%s:%d", m.Host, m.Port)

		var err error
		if m.Port == 465 {
			tlsconfig := &tls.Config{
				InsecureSkipVerify: false,
				ServerName:         m.Host,
			}
			conn, dialErr := tls.Dial("tcp", addr, tlsconfig)
			if dialErr != nil {
				slog.Error("Erro TLS SMTP", "err", dialErr)
				return
			}
			client, clientErr := smtp.NewClient(conn, m.Host)
			if clientErr != nil {
				slog.Error("Erro Client SMTP", "err", clientErr)
				return
			}
			if err = client.Auth(auth); err == nil {
				if err = client.Mail(m.From); err == nil {
					if err = client.Rcpt(to); err == nil {
						w, dataErr := client.Data()
						if dataErr == nil {
							_, _ = w.Write([]byte(message.String()))
							_ = w.Close()
						}
					}
				}
			}
			_ = client.Quit()
		} else {
			err = smtp.SendMail(addr, auth, m.From, []string{to}, []byte(message.String()))
		}

		if err != nil {
			slog.Error("Falha ao enviar e-mail transacional", "to", to, "subject", subject, "err", err)
		} else {
			slog.Info("E-mail transacional enviado com sucesso", "to", to, "subject", subject)
		}
	}()
}

// TemplateBase envelopa a mensagem com estilo visual moderno do Imovxen
func TemplateBase(title string, bodyHTML string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #f8fafc; color: #0f172a; margin: 0; padding: 20px; }
  .card { max-width: 580px; margin: 0 auto; background: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.05); }
  .header { background: linear-gradient(135deg, #2563eb, #1d4ed8); padding: 24px; text-align: center; color: #ffffff; }
  .content { padding: 32px 28px; line-height: 1.6; }
  .btn { display: inline-block; background: #2563eb; color: #ffffff !important; text-decoration: none; padding: 12px 24px; border-radius: 8px; font-weight: bold; margin-top: 16px; }
  .footer { background: #f8fafc; border-top: 1px solid #e2e8f0; padding: 18px; text-align: center; font-size: 12px; color: #64748b; }
</style>
</head>
<body>
  <div class="card">
    <div class="header">
      <h2 style="margin: 0; font-size: 24px; letter-spacing: -0.5px;">Imovxen</h2>
      <p style="margin: 4px 0 0; font-size: 14px; opacity: 0.9;">Marketplace Imobiliário & Gestão de Locação</p>
    </div>
    <div class="content">
      <h3 style="margin-top: 0; font-size: 18px; color: #1e293b;">%s</h3>
      %s
    </div>
    <div class="footer">
      Este é um e-mail transacional gerado automaticamente pela plataforma Imovxen.<br>
      Não responda diretamente a esta mensagem.
    </div>
  </div>
</body>
</html>`, title, bodyHTML)
}

// SendLeadNotification avisa o anunciante sobre um novo cliente interessado
func (m *Mailer) SendLeadNotification(ownerEmail string, ownerName string, propTitle string, leadName string, leadPhone string, leadEmail string, message string) {
	body := fmt.Sprintf(`
	<p>Olá, <strong>%s</strong>!</p>
	<p>Você recebeu uma nova proposta de interesse no seu imóvel <strong>"%s"</strong>.</p>
	<div style="background: #f1f5f9; padding: 16px; border-radius: 8px; margin: 20px 0;">
		<p style="margin: 0 0 8px;"><strong>Interessado:</strong> %s</p>
		<p style="margin: 0 0 8px;"><strong>Telefone / WhatsApp:</strong> %s</p>
		<p style="margin: 0 0 8px;"><strong>E-mail:</strong> %s</p>
		<p style="margin: 0;"><strong>Mensagem:</strong> %s</p>
	</div>
	<p style="margin-bottom: 0;">Acesse seu painel para responder ou entrar em contato direto via WhatsApp.</p>
	<a href="http://localhost:3000/painel/propostas" class="btn">Ver Proposta no Painel</a>
	`, ownerName, propTitle, leadName, leadPhone, leadEmail, message)

	m.Send(ownerEmail, "📬 Nova Proposta Recebida: "+propTitle, TemplateBase("Nova Mensagem de Interessado", body))
}

// SendPaymentConfirmed avisa sobre a compra de anúncios confirmada
func (m *Mailer) SendPaymentConfirmed(userEmail string, userName string, packTitle string, itemsCount int32, amount float64) {
	body := fmt.Sprintf(`
	<p>Olá, <strong>%s</strong>!</p>
	<p>Seu pagamento via Pix de <strong>R$ %.2f</strong> foi confirmado com sucesso!</p>
	<div style="background: #ecfdf5; border: 1px solid #a7f3d0; padding: 16px; border-radius: 8px; margin: 20px 0; color: #065f46;">
		<p style="margin: 0 0 4px; font-weight: bold; font-size: 16px;">🎉 %s</p>
		<p style="margin: 0;">Foram adicionadas <strong>+%d cotas de anúncios ativos</strong> à sua conta.</p>
	</div>
	<p>Você já pode publicar seus novos imóveis no marketplace imediatamente.</p>
	<a href="http://localhost:3000/painel/imoveis" class="btn">Publicar Novo Imóvel</a>
	`, userName, amount, packTitle, itemsCount)

	m.Send(userEmail, "🎉 Pagamento Confirmado: Cotas Liberadas!", TemplateBase("Pagamento Recebido com Sucesso", body))
}

// SendRentalReceipt envia o recibo oficial de quitação para inquilino e proprietário
func (m *Mailer) SendRentalReceipt(toEmail string, recipientName string, propTitle string, installmentNum int32, amount float64, dueDate string, receiptCode string, isLandlord bool) {
	tipo := "Inquilino"
	if isLandlord {
		tipo = "Proprietário"
	}

	body := fmt.Sprintf(`
	<p>Prezado(a) <strong>%s</strong> (%s),</p>
	<p>Confirmamos a <strong>quitação integral</strong> da parcela referente à locação do imóvel <strong>"%s"</strong>.</p>
	<div style="background: #ffffff; border: 2px dashed #cbd5e1; padding: 20px; border-radius: 8px; margin: 20px 0;">
		<div style="font-size: 13px; text-transform: uppercase; color: #64748b; font-weight: bold; letter-spacing: 1px;">Recibo Oficial de Quitação de Aluguel</div>
		<div style="font-size: 22px; font-weight: bold; color: #16a34a; margin: 8px 0 16px;">R$ %.2f &bull; QUITADO</div>
		<p style="margin: 0 0 6px;"><strong>Parcela:</strong> %d</p>
		<p style="margin: 0 0 6px;"><strong>Vencimento:</strong> %s</p>
		<p style="margin: 0 0 6px;"><strong>Código de Autenticidade:</strong> <code style="background:#f1f5f9; padding:2px 6px; border-radius:4px;">%s</code></p>
		<p style="margin: 0;"><strong>Status:</strong> Liquidado com validade jurídica via Imovxen</p>
	</div>
	<p>Este comprovante atesta a regularidade do pagamento para todos os fins de direito.</p>
	`, recipientName, tipo, propTitle, amount, installmentNum, dueDate, receiptCode)

	m.Send(toEmail, "📄 Recibo de Aluguel Quitado - Parcela "+strconv.Itoa(int(installmentNum))+" ("+propTitle+")", TemplateBase("Recibo de Quitação de Aluguel", body))
}
