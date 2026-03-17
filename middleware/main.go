package main

func main() {
	db := connectDB()
	defer db.Close()

	repo := NewPostgresRepository(db)
	consumeMessages(repo)
}