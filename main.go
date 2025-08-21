package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Activity struct {
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Category     string    `json:"category"`
	Description  string    `json:"description"`
	ActivityDate time.Time `json:"activity_date"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func initDB() (*sql.DB, error) {
	godotenv.Load()
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	dns := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s", user, password, host, port, dbname)
	db, err := sql.Open("postgres", dns)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	app := fiber.New()

	// Get /activities
	app.Get("/activities", func(c *fiber.Ctx) error {
		rows, err := db.Query("SELECT * FROM activities")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}
		defer rows.Close()

		var activities []Activity
		for rows.Next() {
			var activity Activity
			err = rows.Scan(&activity.ID, &activity.Title, &activity.Category, &activity.Description, &activity.ActivityDate, &activity.Status, &activity.CreatedAt)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"message": err.Error(),
				})
			}
			activities = append(activities, activity)
		}
		return c.Status(fiber.StatusCreated).JSON(activities)
	})

	app.Listen(":8081")
}
