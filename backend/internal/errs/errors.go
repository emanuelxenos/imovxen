package errs

import "errors"

var (
	ErrNotFound          = errors.New("recurso não encontrado")
	ErrUnauthorized      = errors.New("não autorizado")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrEmailAlreadyExists = errors.New("e-mail já cadastrado")
	ErrQuotaExceeded     = errors.New("limite de cota de anúncios atingido")
	ErrInvalidToken      = errors.New("token inválido ou expirado")
	ErrForbidden         = errors.New("acesso negado ao recurso")
)
