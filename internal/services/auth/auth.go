package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/lib/logger/sl"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)


type Auth struct {
	log *slog.Logger
	usrSaver UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAppId = errors.New("invalid app id")
	ErrUserExists = errors.New("user already exists")
	ErrUserNotFound = errors.New("user not found")
)


// New returns a new instance of the Auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log: log,
		usrSaver: userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL: tokenTTL,
	}
}


func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,	
) (string, error) {

	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", sl.Err(err))
			// Тут принципиально не стоит говорить человеку о том что он ввел неправильно -
			// либо он ввел логин неправильно, или он ввел пароль не правильно, например таким -
			// образом человек не сможет составить базу тех людей которые у нас зарегестрированы 
			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", sl.Err(err))
		
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", sl.Err(err))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	// Получаем приложение в которое пользователь хочет залогиниться - 
	// мы идем в a.appProvider.App(ctx, appID) получаем по id этого приложения информацию -
	// о нем потому что следующим этапом мы будем создавать токен, и для использования этого -
	// токена мы используем ключ конкретно этого приложения.
	// Каждый токен подписывается ключом, тоесть у каждого приложения есть свой собственный ключ - 
	// потму что возможно SSO пользуются разные группы разработчиков, и нельзя между ними делиться -
	// одним общим ключом.
	// Обычно для каждого отдельного приложения используется свой собственный ключ, и в зависимости -
	// от алгоритма по которому будет кодироваться JWT. В нашем случае приватный ключ будет храниться и -
	// на стороне клиентского сервиса и на стороне SSO, соответственно для того чтобы понять каким ключом -
	// его подписывать нужно узнать что это за приложение, и секретный ключ этого приложения у нас будет -
	// хранится в базе данных (что не очень хорошо)

	app, err := a.appProvider.App(ctx, appID)
	if err != nil{
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")
	
	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to generate token", sl.Err(err))
		
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}


	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", sl.Err(err))
	
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			log.Warn("user already exists", sl.Err(err))
			
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed to save user", sl.Err(err))

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("user registered")

	return id, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),  
		slog.String("user_id", string(userID)),
	)

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		// Мы не пробрасываем дальше ошибку (ErrAppNotFound), потому что это ошибка -
		// слоя работы с данными, это не ошибка сервисного слоя, и про ошпибку (ErrAppNotFound) - 
		// хендлер не должен занть, хендлер может знать про ошибку ErrInvalidAppId
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", sl.Err(err))

			return false, fmt.Errorf("%s: %w", op, err)
		}
		return false, fmt.Errorf("%s: %w", op, ErrInvalidAppId)
	}
	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))


	// panic("not implemented")
	return isAdmin, nil
}