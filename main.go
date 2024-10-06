package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type Headers struct {
	UserAgent          string `json:"user_agent"`
	AuthorizationToken string `json:"authorization"`
	ContentType        string `json:"content_type"`
}

type Request struct {
	Headers Headers `json:"headers"`
	Body    string  `json:"body"`
}

// Connecting to database
func initDB(log *logrus.Logger) (*sql.DB, error) {
	log.Info("Connection to database")
	connStr := fmt.Sprintf("postgres://%s:%s@postgres-db:5432/%s?sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	// Test connection
	if err = db.Ping(); err != nil {
		return nil, err
	}
	// Init table
	createTable := "CREATE TABLE event_users (id SERIAL PRIMARY KEY, time TIMESTAMP DEFAULT CURRENT_TIMESTAMP, user_id TEXT NOT NULL, data TEXT NOT NULL);"
	if _, err = db.Exec(createTable); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P07" {
			log.Info("Table already exists")
		} else {
			log.Error(err)
		}
	} else {
		log.Info("Table create successful")
	}
	return db, nil
}

func handleRequest(c *fiber.Ctx, ch chan *Request, log *logrus.Logger) error {
	// Parse request
	headers := c.GetReqHeaders() // map[string][]string
	requiredHeaders := map[string]*string{
		"X-Tantum-Useragent":     nil,
		"X-Tantum-Authorization": nil,
		"Content-Type":           nil,
	}

	for key := range requiredHeaders {
		if value, ok := headers[key]; ok && len(value) > 0 {
			requiredHeaders[key] = &value[0]
		} else {
			log.Error("Missing required headers: ", key)
			return c.Status(400).JSON(fiber.Map{
				"message": "missing required headers",
			})
		}
	}

	request := Request{
		Headers: Headers{
			UserAgent:          *requiredHeaders["X-Tantum-Useragent"],
			AuthorizationToken: *requiredHeaders["X-Tantum-Authorization"],
			ContentType:        *requiredHeaders["Content-Type"],
		},
		Body: string(c.BodyRaw()),
	}

	ch <- &request

	// Response
	c.Status(202).JSON(fiber.Map{
		"status": "ok",
	})

	return nil
}

// Init logger
func newLogger(logger *logrus.Logger, level string) *logrus.Logger {
	logger.SetFormatter(&logrus.TextFormatter{})
	switch level {
	case "FATAL":
		logger.SetLevel(logrus.FatalLevel)
	case "INFO":
		logger.SetLevel(logrus.InfoLevel)
	case "ERROR":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.DebugLevel)
	}
	return logger
}

func main() {
	log := newLogger(logrus.New(), os.Getenv("LOG_LEVEL"))

	app := fiber.New()

	db, err := initDB(log)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create worker pool
	log.Info("Create worker pool")
	poolCount, err := strconv.Atoi(os.Getenv("POOL_COUNT"))
	if err != nil {
		log.Error("Error env POOL_COUNT")
	}
	ch := make(chan *Request, poolCount)

	for i := 0; i < poolCount; i++ {
		go func(ch chan *Request) {
			for request := range ch {
				head, err := json.Marshal(request.Headers)
				if err != nil {
					log.Error("Error marshaling request headers: ", err)
					continue
				}
				log.Info("Insert into database")
				_, err = db.Exec("INSERT INTO event_users (user_id, data) VALUES ($1,$2)", string(head), request.Body)
				if err != nil {
					log.Error("Error inserting into database: ", err)
				}
				log.Info("The insertion was successful")
			}
		}(ch)
	}
	log.Info("Worker pool created")

	// Handlers
	app.Post("/analytics", func(c *fiber.Ctx) error {
		log.Info(fmt.Sprintf("Request processed. Method:%s Route:%s", c.Method(), c.Path()))
		return handleRequest(c, ch, log)
	})
	app.Get("/metrics", monitor.New())

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
