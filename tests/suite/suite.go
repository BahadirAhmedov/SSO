package suite

import (
	"context"
	"net"
	"sso/internal/config"
	"strconv"
	"testing"

	ssov1 "github.com/BahadirAhmedov/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)


type Suite struct{
	*testing.T	// Потребуется для вызова методов *testing.T внутри Suite
	Cfg *config.Config // Конфигурация приложения
	AuthClient ssov1.AuthClient // Клиент для взаимодействия с gRPC-сервером
}

const(
	grpcHost = "localhost"
)

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper() // Нужна для того чтобы при фэйле какого-то теста у нас правильно формировался -
	// стэк вызовов и эта функция не была указана как финальная 
	t.Parallel() // Указываем что мы можем выполнять наши тесты паралельно (сильно ускоряеет работу -
	// но при написании пралельных тестов нужно учитывать некоторые нюансы )

	cfg := config.MustLoadByPath("../config/local.yaml")


	// Создаем context для того чтобы передавать во все дочерние функции, он нужен как минимум для того чтобы -
	// тесты слишком сильно не затягивались 
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)

	// Этот context мы будем отменять в тот момент когда закончятся тесты, тоесть мы вызываем функцию Cleanup() -
	// она будет выполняться когда тест полностью закончен, чтобы тест за собой что-нибудь почистил 

	t.Cleanup(func() {
		t.Helper()
		cancelCtx() // отменяем context
	})


	// Создаем grpc клиента для нашего сервиса 
	cc, err := grpc.DialContext(context.Background(),
		grpcAdress(cfg),
		// указываем что мы будем использовать, insecure.NewCredentials()
		// insecure.NewCredentials() - редоставляет "пустые" креденшлы, которые отключают TLS (будем -
		// использовать небезопасное соединение)
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}	


	return ctx, &Suite{
		T: t,
		Cfg: cfg,
		// Создаем Auth клиент (написан заранее при кодо генерации)  
		AuthClient: ssov1.NewAuthClient(cc),
	}
}


func grpcAdress(cfg *config.Config) string {
// функция JoinHostPort объеденяет хост и порт в общий адрес
return 	net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}


