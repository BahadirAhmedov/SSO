package tests

import (
	"fmt"
	"sso/tests/suite"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	ssov1 "github.com/BahadirAhmedov/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const(
	emptyAppID = 0
	appID = 1 // Это тот который мы подготовили заранее в тестовых миграциях 
	appSecret = "test-secret"

	passDefaultLen = 10
)


func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	
	email := gofakeit.Email()
	pass := randomFakePassword()

	// Используем клиент 
	// Хоть мы регистрируем здесь логин пользователя, мы все же первым делом выполним регистрацию - 
	// потому что мы не надеемся ни на какие предыдущие данные, мы сначала их подготовим, подготовим -
	// по честному, если мы хотим логинить человека, мы сначала его зарегестрируем 
	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email: email,
		Password: pass, 
	})

	// Если мы не выполнили проверку require то тест фейлится и дальше не идет, если мы не смогли создать -
	// клиента то нет смысла что-то дальше тестировать
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email: email,
		Password: pass,
		AppId: appID,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	require.NotEmpty(t, token)


	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)


	claims, ok := tokenParsed.Claims.(jwt.MapClaims)	
	require.True(t, ok)
	// Проверяем что в токене содержится коректная информация 
	// Сверяем id, мы берем из ответа который пришел при регистрации respReg, и проверяем -
	// что именно он лежит, также и в токене 
	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	// Проверяем email который мы сгенерировали для регистрации
	assert.Equal(t, email, claims["email"].(string))
	// И appID который мы определили в константе и миграции
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	// Дальше нужно проверить что время истичения токена совпадает с ожидаемым, но точно это -
	// проверить не сможем потому что мы не знаем в какой момент времени до секунды происходила -
	// генерация токена, потому что приложение работает само по себе, а тест сам по себе, но мы можем -
	// примерно прикинуть, прикинем с точностью до 1 секунды. 
	// Примерно момент когда произошла генерация токена это, сразу же после того как мы вызвали логин -
	// тоесть, когда мы вызываем логин в этот момент создается токен в нем сразу вписывается время его жизни -
	// и сразу после этого мы засикаем время loginTime := time.Now()
	
	const deltaSeconds = 1
	// Берем время логина добавляем к нему duration из конфига, и проверяем что в токене лежит то же самое значение -
	// с точностью до 1 секунды 
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

// Фэйл кейс
// Пользователь может попытаться дважды зарегестрироваться 

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)
	
	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email: email,
		Password: pass, 
	})

	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email: email,
		Password: pass, 
	})

	fmt.Println("err)):", err)

	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}


// Табличные кейсы, потомучто все проверки будут примерно одинаковыми, отличатся будут только -
// входные параметры

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name string
		email string
		password string
		expectedErr string
	}{
		{
			name: "Register with Empty Password",
			email: gofakeit.Email(),
			password: "",
			expectedErr: "password is required",
		},
		{
			name: "Register with Empty Email",
			email: "",
			password: randomFakePassword(),
			expectedErr: "email is required",
		},
		{
			name: "Register with Both Empty",
			email: "",
			password: "",
			expectedErr: "email is required",
		},
	}
	// Мы просто перечисляем различные наборы параметров, которые в цикле будут по очареди выполнены - 
	for _, tt := range tests {
		// Можно модефицировать этот блок таким образом чтобы наши тесты запускались внутри цикла паралельно
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email: tt.email,
				Password: tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}



}
func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		email       string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			email:       gofakeit.Email(),
			password:    "",
			appID:       appID,
			expectedErr: "password is required",
		},
		{
			name:        "Login with Empty Email",
			email:       "",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "Login with Both Empty Email and Password",
			email:       "",
			password:    "",
			appID:       appID,
			expectedErr: "email is required",
		},
		{
			name:        "Login with Non-Matching Password",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid credentials",
		},
		{
			name:        "Login without AppID",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appID:       emptyAppID,
			expectedErr: "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Соаздаем пользователя 
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				// Мы здесь не используем, данные из кейса, мы используем случайные данные - 
				// а в кейсе случайные другие данные 
				Email:    gofakeit.Email(),
				Password: randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tt.email,
				Password: tt.password,
				AppId:    tt.appID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}