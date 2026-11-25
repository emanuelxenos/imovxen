# TASK-006: Frontend SSR com Templ, HTMX e Cliente API

## Objetivo
Estruturar o projeto `frontend/` em Go com Gin, configurar a compilação do `templ`, incluir HTMX e assets, e implementar o cliente HTTP interno com renovação de sessão.

## Dependências
- `TASK-004-backend-properties-leads.md`
- `documento.md` (Seção 3, 4 e 6)

## Entregáveis
1. Inicialização do módulo Go em `frontend/go.mod`.
2. Servidor Gin em `frontend/cmd/web/main.go` servindo estáticos (`static/`).
3. Assets em `frontend/static/`:
   - `htmx.min.js`.
   - Folha de estilo base (Tailwind compilado ou CSS customizado moderno).
4. Módulo `frontend/internal/api/client.go`:
   - Cliente HTTP que repassa cookie/token para o backend.
   - Interceptador de 401 para auto-refresh silencioso.
5. Layout base em `frontend/internal/views/layouts/base.templ`:
   - Header, Navbar com estado de login, container principal, Footer.
6. Páginas da Fase 1:
   - Home (`/`) com barra de busca e listagem de cards via HTMX.
   - Detalhe do Imóvel (`/imoveis/:id`) com galeria e formulário de contato.
   - Auth (`/login` e `/cadastro`).
   - Painel do Usuário (`/painel/imoveis`) com formulário de novo anúncio e upload de fotos.

## Critério de Aceite
- O frontend compila e renderiza páginas SSR rapidamente.
- Filtro de busca na Home atualiza a lista de imóveis via HTMX sem recarregar a página inteira.
- O fluxo de cadastro, login, publicação de anúncio e envio de contato funciona de ponta a ponta pelo navegador.
