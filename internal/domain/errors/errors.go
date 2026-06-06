package domainerrors

import "errors"

var (
	// ErrUserAlreadyExists é retornado quando um usuário com o mesmo username já existe.
	ErrUserAlreadyExists = errors.New("user with this username already exists")

	// ErrInvalidCredentials é retornado quando username ou senha estão incorretos.
	ErrInvalidCredentials = errors.New("invalid username or password")

	// ErrQuizNotFound é retornado quando o quiz não é encontrado.
	ErrQuizNotFound = errors.New("quiz history not found")

	// ErrQuizForbidden é retornado quando o usuário não tem permissão para a operação.
	ErrQuizForbidden = errors.New("you are not allowed to perform this action")

	// ErrQuizAlreadyExists é retornado quando o usuário já realizou um quiz com o mesmo tema e dificuldade.
	ErrQuizAlreadyExists = errors.New("quiz with this topic and difficulty already exists for this user")

	// ErrInvalidTopic é retornado quando o tema não passa nas validações de guard rail.
	ErrInvalidTopic = errors.New("invalid topic: use a legitimate educational subject")
)
