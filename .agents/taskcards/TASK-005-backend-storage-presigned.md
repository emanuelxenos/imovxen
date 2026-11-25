# TASK-005: Upload de Imagens Direto com Presigned URLs (S3/MinIO)

## Objetivo
Implementar o serviço de geração de URLs pré-assinadas no backend para permitir upload direto do navegador ao storage, sem sobrecarregar a API.

## Dependências
- `TASK-004-backend-properties-leads.md`
- `documento.md` (Seção 4.3 e 6.3)

## Entregáveis
1. Módulo `backend/internal/storage/`:
   - Cliente S3 (usando AWS SDK v2 para Go) compatível com MinIO e Cloudflare R2.
   - Geração de Presigned PUT URL com validade de 5 minutos, chave única baseada em UUID e validação de MIME type permitido.
2. Endpoints:
   - `POST /api/properties/:id/images/presign` (gera a URL assinada).
   - `POST /api/properties/:id/images/confirm` (salva a referência em `property_images` e define `cover_image_url` se for capa).
   - `DELETE /api/properties/:id/images/:img_id` (remove o registro e o objeto no storage).

## Critério de Aceite
- O endpoint de presign retorna uma URL válida capaz de receber um arquivo via `PUT` HTTP.
- Ao confirmar o upload, a imagem é persistida no banco e associada ao imóvel correto.
