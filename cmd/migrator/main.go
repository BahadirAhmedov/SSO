package main

import (
	"errors"
	"flag"
	"fmt"
	
	// Библиотека для миграций
	"github.com/golang-migrate/migrate/v4"

	// Драйвер для выполнения миграций SQLite 3
	// Подключается драйвер для работы с SQLite 3, это не тот драйвер -
	// который мы подключаем для работы с SQLite со своим собственным -
	// приложением
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	// Драйвер для получения миграций из файлов
	_ "github.com/golang-migrate/migrate/v4/source/file"

)

func main() {
	var storagePath, migrationsPath, migrationsTable string
	// storagePath - путь нашего хранилища, то есть базы данных в которой -
	// нужно применить миграции 
	flag.StringVar(&storagePath, "storage-path", "", "path to storage")

	// migrationsPath - тоесть где у нас находятся файлы миграции 
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	
	// migrationsTable - используется для того чтобы имя таблицы в которую мигратор будет -
	// сохранять изменеия, тоесть мигратору чтобы понимать какие нужно применить миграции -
	// ему нужно посмотреть какие уже применены ранее, соответственно когда мы будем что то -
	// откатывать это информация также потребуется и таблицу снова придется модифицировать -
	// и если этот параметр не задавать то у таблицы будет стандартное имя migrations, и 
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations")
	flag.Parse()

	if storagePath == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migratiosn-path is required")
	}
	
	// Создаем экземпляр мигратора, тоесть объект который будет выполнять миграции (и здесь от стандартного, немножко отличается )
	// ?x-migrations-table= - указываем отдельный параметр который отвечает за именование таблицы в которой будет храниться -
	// информация о миграциях, если этим параметром пользоваться не будут то можно этот параметр удалить(?x-migrations-table=)
	m, err := migrate.New(
		"file://" + migrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)
	if err != nil {
		panic(err)
	}

	// Выполняем саму миграцию, за это отвечает метод Up, он выполнит все недостающии миграции вплоть до самой
	if err := m.Up(); err != nil {
		// ErrNoChange - это ошибка будет возвращаться в том случае если все миграции уже применены, и применять -
		// нечего в данном случае это не ошибка, просто, напишем в консоль что нечего мигрировать 
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
		
			return
		}
		panic(err)
	}
	fmt.Println("migrations applied successfully")
}