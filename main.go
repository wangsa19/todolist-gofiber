package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Activity struct {
	ID           int       `json:"id"`
	Title        string    `json:"title" validate:"required"`
	Category     string    `json:"category" validate:"required,oneof=TASK EVENT"`
	Description  string    `json:"description" validate:"required"`
	ActivityDate time.Time `json:"activity_date" validate:"required"`
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
	validate := validator.New()

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

	// Post /activities
	app.Post("/activities", func(c *fiber.Ctx) error {
		var activity Activity
		err := c.BodyParser(&activity)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		if err = validate.Struct(&activity); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		sqlStatement := `INSERT INTO activities(title, category, description, activity_date, status) VALUES($1, $2, $3, $4, $5) RETURNING id`
		err = db.QueryRow(sqlStatement, activity.Title, activity.Category, activity.Description, activity.ActivityDate, "NEW").Scan(&activity.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success"})
	})

	// PUT /activities/:id
	app.Put("/activities/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var activity Activity
		err := c.BodyParser(&activity)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		if err = validate.Struct(&activity); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		sqlStatement := `UPDATE activities SET title=$1, category=$2, description=$3, activity_date=$4 WHERE id=$5 RETURNING id`
		err = db.QueryRow(sqlStatement, activity.Title, activity.Category, activity.Description, activity.ActivityDate, id).Scan(&activity.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	})

	// DELETE /activities/:id
	app.Delete("/activities/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		sqlStatement := `DELETE FROM activities WHERE id=$1`
		_, err := db.Exec(sqlStatement, id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	})

	app.Listen(":8081")
}
