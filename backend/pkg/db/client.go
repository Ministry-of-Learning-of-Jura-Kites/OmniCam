package db_client

// func Migrate() {
// 	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
// 	m, err := migrate.NewWithDatabaseInstance(
// 		"file:///migrations",
// 		"postgres", driver)

// 	driver, err := postgres.WithInstance(pool, &postgres.Config{})
// 	m, err := migrate.NewWithDatabaseInstance(
// 		"file:///migrations",
// 		"postgres", driver)
// 	m.Up() // or m.Steps(2) if you want to explicitly set the number of migrations to run
// }
