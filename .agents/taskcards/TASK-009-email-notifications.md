# TASK-009: Serviço de E-mails Transacionais Automáticos

## Contexto
O sistema precisa manter anunciantes, proprietários e inquilinos informados em tempo real através de e-mails profissionais automáticos.
Este módulo implementa o serviço de e-mail transacional no backend, com suporte flexível a SMTP padrão (Gmail, Locaweb, Hostgator, etc.) e provedores modernos como **Resend** / **SendGrid**, além de um modo de log/simulação transparente para ambiente local de desenvolvimento.

## Eventos com Disparo Automático de E-mail
1. **Nova Proposta / Lead Recebido**:
   - Destinatário: Proprietário / Corretor do imóvel.
   - Conteúdo: Nome do interessado, telefone/WhatsApp, mensagem enviada e link direto para responder.
2. **Confirmação de Pagamento de Cota de Anúncio**:
   - Destinatário: Usuário anunciante.
   - Conteúdo: Confirmação da liquidação do Pix, pacote adquirido e novas cotas disponíveis na conta.
3. **Novo Contrato de Locação & Carnê Gerado**:
   - Destinatário: Inquilino (e cópia para o proprietário).
   - Conteúdo: Boas-vindas, resumo do aluguel (valor mensal, dia do vencimento) e aviso de que as faturas do carnê estão ativas.
4. **Recibo de Quitação de Aluguel (Baixa Manual ou Sistema)**:
   - Destinatário: Inquilino e Proprietário.
   - Conteúdo: Recibo com número da parcela, mês de referência, valor quitado e código de validação legal (`REC-...`).

## Dependências
- `backend/internal/mailer/`
- Pacote nativo `net/smtp` ou integração HTTP Resend.

## Entregáveis
- Serviço `mailer.Service` com templates HTML responsivos com a identidade visual do Imovxen.
- Chamadas assíncronas (via goroutines Go `go mailer.Send...()`) nos endpoints de leads, pagamentos e contratos, garantindo resposta de 0 ms sem travar a interface do usuário.
- Variáveis de ambiente configuráveis em `.env`: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `MAIL_FROM`.

## Critérios de Aceite
- Ao enviar proposta na página do imóvel, o proprietário recebe a notificação por e-mail automaticamente.
- Ao dar baixa manual ou confirmar Pix, inquilino e proprietário recebem o e-mail de quitação/recibo.
- Falhas de envio de e-mail não travam ou revertem transações do banco de dados (envio em background com logging).
