package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/hietkamp/norma-wrk/internal/handlers"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

const (
	Version = "v1.0.1"
	Banner  = `

╔═╗ ╔╗                
║║╚╗║║                
║╔╗╚╝║╔══╗╔═╗╔╗╔╗╔══╗ 
║║╚╗║║║╔╗║║╔╝║╚╝║╚ ╗║ 
║║ ║║║║╚╝║║║ ║║║║║╚╝╚╗
╚╝ ╚═╝╚══╝╚╝ ╚╩╩╝╚═══╝ %s
Norma Worker, %s

___________________________________________________________/\__/\__0>______

`
)

type Server struct {
	Reader *kafka.Reader
}

var server *Server

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal().Msg("Error loading .env file")
	}
	brokers := strings.Split(os.Getenv("kafka_url"), ",")
	server = &Server{}
	server.Reader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		GroupID:  os.Getenv("group_id"),
		Topic:    os.Getenv("topic_reader"),
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	defer server.Reader.Close()

	fmt.Printf(Banner, string(colorRed)+Version+string(colorReset), string(colorCyan)+"https://infosupport.com"+string(colorReset))
	fmt.Print("\n")

	log.Info().Msg("Consumer starts reading ...")
	for {
		m, err := server.Reader.FetchMessage(ctx)
		if err == io.EOF {
			return
		}
		if err == nil {
			if m.Value != nil {
				err = handlers.HandleValidatedQueryReceived(m.Value)
				if err != nil {
					log.Error().Msgf("error handling message: %w", err)
				}
			}
			// Commit the message if the handling is done
			if err := server.Reader.CommitMessages(ctx, m); err != nil {
				log.Error().Msgf("failed to commit messages:", err)
			}
		}
		// Listen for the interrupt signal.
		<-ctx.Done()
		break
	}
	log.Info().Msgf("Server exiting\n")
}
