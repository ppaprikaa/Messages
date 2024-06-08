package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

type MessageReq struct {
	Message string `json:"message"`
}

type MessageRes struct {
	ID   int    `json:"id"`
	Text string `json:"name"`
}

type ErrorRes struct {
	Message string `json:"message"`
}

func main() {
	cfg := readConf(os.Getenv("MESSAGES_CONF"))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err = conn.Exec(ctx, `create table if not exists messages (
		id SERIAL PRIMARY KEY,
		message TEXT NOT NULL
	)`); err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	e.POST("/api/v1/message", func(c echo.Context) error {
		messageBody := new(MessageReq)

		if err := c.Bind(messageBody); err != nil {
			return err
		}

		if !ValidateMessage(messageBody.Message) {
			message := "Input should:"
			if len([]rune(messageBody.Message)) < 20 {
				message += " Have length > 20 characters."
			}
			if len([]rune(messageBody.Message)) > 0 {
				if !unicode.IsUpper([]rune(strings.TrimSpace(messageBody.Message))[0]) {
					message += " First character must be uppercase."
				}
			}

			message += " done."

			return c.JSON(http.StatusUnprocessableEntity, ErrorRes{Message: message})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		row := conn.QueryRow(
			ctx,
			"insert into messages (message) values ($1) returning id",
			messageBody.Message,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorRes{Message: err.Error()})
		}

		var id int

		if err = row.Scan(&id); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorRes{Message: err.Error()})
		}

		return c.JSON(http.StatusCreated, MessageRes{ID: id, Text: messageBody.Message})
	})

	e.GET("/api/v1/message/:id", func(c echo.Context) error {
		idParam := c.Param("id")

		id, err := strconv.Atoi(idParam)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorRes{Message: fmt.Sprintf(`server internal error: failed to convert "%s" to int`, idParam)})
		}

		query := `select id, message from messages where id=$1`
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		row := conn.QueryRow(ctx, query, id)

		var dstMsg string
		var dstID int

		if err := row.Scan(&dstID, &dstMsg); err != nil {
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				return c.JSON(http.StatusNotFound, ErrorRes{Message: fmt.Sprintf("message with id %d not found", id)})
			default:
				return c.JSON(http.StatusInternalServerError, ErrorRes{Message: "internal server error: failed to select message from db"})
			}
		}

		return c.JSON(http.StatusOK, MessageRes{ID: dstID, Text: dstMsg})
	})

	e.Logger.Fatal(e.Start(":8080"))
}

func ValidateMessage(message string) bool {
	if len([]rune(message)) < 20 || !unicode.IsUpper([]rune(strings.TrimSpace(message))[0]) {
		return false
	}

	return true
}

type Config struct {
	DSN string `json:"dsn" yaml:"dsn" toml:"dsn"`
}

func readConf(filename string) *Config {
	var cfg Config

	if len(strings.TrimSpace(filename)) == 0 {
		filename = "config.local.yaml"
	}

	if err := cleanenv.ReadConfig(filename, &cfg); err != nil {
		log.Fatal(err)
	}

	return &cfg
}
