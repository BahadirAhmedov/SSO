package jwt

import (
	"sso/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// покрыть функцию тестами задание
func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// Сформируем объект в котором будут храниться все данные которые мы будем дальше передовать -
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()
	claims["app_id"] = app.ID
	
	// Секрет хранить в модели не очень хорошая идея, потому что эта модель имеет риск быть залогированной - 
	// а в логи всякие секреты нельзя помещать 
	tokenString, err := token.SignedString([]byte(app.Secret))

	if err != nil {
		return "", err
	}

	return tokenString, nil 
}