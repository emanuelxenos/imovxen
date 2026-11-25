---
name: imovxen-frontend-guide
description: >-
  Guia para desenvolvimento no frontend Go SSR do Imovxen. Use sempre que for criar
  páginas, componentes templ, interações HTMX ou handlers de visualização.
---

# Frontend SSR & HTMX Guide — Imovxen

Este guia descreve os padrões para o frontend em `frontend/` utilizando Go, Gin, `a-h/templ` e HTMX.

## 1. Arquitetura de Visualização com `templ`

1. **Separação de Componentes**:
   - `internal/views/layouts/`: Estrutura HTML base (Navbar, Footer, import do HTMX, CSS, meta tags).
   - `internal/views/pages/`: Páginas completas (ex: Home, Detalhe do Imóvel, Dashboard do Usuário).
   - `internal/views/partials/`: Fragmentos HTML reativos para respostas HTMX (ex: cards de imóveis, modal de contato, lista filtrada).
2. **Compilação**:
   - Componentes `.templ` são compilados para código Go através de `templ generate`.

## 2. Consumo da API Backend (`internal/api`)

- O frontend **NUNCA** acessa o banco de dados.
- Toda interação com os dados passa pelo cliente HTTP em `internal/api/client.go`.
- **Auto-Refresh de Token**:
  - O cliente anexa o token JWT recebido nos cookies do usuário.
  - Se a API retornar `401 Unauthorized`, o cliente tenta silenciosamente renovar o token chamando `POST /api/auth/refresh` com o Refresh Token e repete a chamada.

## 3. Padrões de Interatividade HTMX

- **Busca e Filtros com Delay**:
  ```html
  <input type="text" name="city" 
         hx-get="/imoveis/partials/lista" 
         hx-trigger="keyup changed delay:400ms" 
         hx-target="#lista-imoveis" 
         hx-indicator="#loading-spinner" />
  ```
- **Paginação / Infinite Scroll**:
  - Use `hx-swap="beforeend"` no container de listagem para carregar novos itens de forma fluida.
