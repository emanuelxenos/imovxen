# TASK-001: Infraestrutura Básica e Ambiente Docker Local

## Objetivo
Configurar o ambiente Docker local com PostgreSQL 16 (com extensões ativas) e MinIO para simulação de storage S3 local.

## Dependências
- `documento.md` (Seção 2, 5 e 9)

## Entregáveis
1. Arquivo `docker-compose.yml` na raiz:
   - Serviço `postgres`: imagem `postgres:16-alpine`, banco `imovxen`, porta `5432`.
   - Serviço `minio`: imagem `minio/minio`, portas `9000` (API) e `9001` (Console).
   - Serviço `minio-create-bucket`: container temporário para criar automaticamente o bucket `imovxen-images` no MinIO ao iniciar.
   - Volumes persistentes para dados do Postgres e MinIO.
2. Arquivo `.env.example` com todas as variáveis documentadas.

## Critério de Aceite
- Executar `docker compose up -d` sobe os serviços sem erros.
- É possível conectar no PostgreSQL e MinIO nas portas locais.
