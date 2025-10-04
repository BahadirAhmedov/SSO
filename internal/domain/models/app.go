package models

type App struct {
	ID int
	Name string
	// Секрет будет использоваться для того чтобы подписывать токены - 
	// и потом на стороне клиентского приложения эти токены валидировать
	Secret string
}