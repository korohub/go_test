package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/template/jet/v2"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

// Database instance
var db *sql.DB

// Database settings
const (
	host     = "localhost"
	port     = 5432 // Default port
	user     = "postgres"
	password = "jodagrens"
	dbname   = "onairjob"
)

// job struct
type Job struct {
	// ID     string `json:"id"`
	Job   string `json:"job"`
}

// Jobs struct
type Jobs struct {
	Jobs []Job `json:"jobs"`
}

// Connect function
func Connect() error {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

type book struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

var books = []book{
	{ID: "1", Title: "In Search of Lost Time", Author: "Marcel Proust", Quantity: 2},
	{ID: "2", Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", Quantity: 5},
	{ID: "3", Title: "War and Peace", Author: "Leo Tolstoy", Quantity: 6},
}

func main() {

	// Connect with database
	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	// Create a new engine
	engine := jet.New("./views", ".jet")

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "jodagrens", // no password set
		DB:       0,           // use default DB
	})

	ctx := context.Background()

	err := client.Set(ctx, "hello", "Hello, World from REDIS!", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, "hello").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("hello", val)

	// Or from an embedded system
	// See github.com/gofiber/embed for examples
	// engine := jet.NewFileSystem(http.Dir("./views", ".jet"))

	// Pass the engine to the views
	app := fiber.New(fiber.Config{
		Views: engine,
		// Prefork: true,
	})
	//  compress Brotli always before static and route
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))

	app.Static("/static", "./static")

	app.Use(favicon.New())

	// Or extend your config for customization
	app.Use(favicon.New(favicon.Config{
		File: "./static/favicon.ico",
		URL:  "/favicon.ico",
	}))

	app.Get("/metrics", monitor.New(monitor.Config{Title: "MyService Metrics Page"}))

	// app.Get("/", func(c *fiber.Ctx) error {
	// 	// Render index

	// 	return c.Render("index", fiber.Map{
	// 		"Title": "Hello, World!",
	// 	}, "layouts/main")
	// })
	app.Get("/", func(c *fiber.Ctx) error {
		// Render index

		return c.Render("index", fiber.Map{
			"Title": "Hello, World!",
		}, "layouts/main")
	})

	app.Get("/job", func(c *fiber.Ctx) error {
		// Select all Employee(s) from database
		rows, err := db.Query("SELECT job FROM job limit 10")
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()
		result := Jobs{}

		for rows.Next() {
			job := Job{}
			if err := rows.Scan(&job.Job); err != nil {
				return err // Exit if we get an error
			}

			// Append job to jobs
			result.Jobs = append(result.Jobs, job)
		}
		// Return Employees in JSON format
		return c.JSON(result)
	})

	// Simple GET handler
	app.Get("/api/list", func(c *fiber.Ctx) error {
		// fmt.Println(c.Request())
		return c.SendString("I'm a GET request!")
	})

	// Simple GET handler
	app.Get("/api/redis", func(c *fiber.Ctx) error {
		val, err := client.Get(ctx, "hello").Result()
		if err != nil {
			panic(err)
		}
		return c.SendString(val)
	})
	type SomeStruct struct {
		Name string
		Age  uint8
	}

	// data := SomeStruct{
	// 	Name: "Grame",
	// 	Age:  20,
	// }

	app.Get("/books", func(c *fiber.Ctx) error {
		return c.JSON(books)
	})

	app.Get("/json", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name": "Grame",
			"age":  20,
		})
	})

	app.Post("/body", func(c *fiber.Ctx) error {
		// Get raw body from POST request:
		return c.Send(c.BodyRaw()) // []byte("user=john")
	})

	log.Fatal(app.Listen(":3045"))
}
