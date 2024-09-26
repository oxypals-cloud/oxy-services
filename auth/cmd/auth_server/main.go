package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"oxypals-cloud/oxy-services/auth/internal/server"
	"oxypals-cloud/oxy-services/auth/proto/github.com/oxypals-cloud/oxy-services/auth"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
)

var (
	dbURI     string
	secretKey string
	port      string
)

func init() {
	flag.StringVar(&dbURI, "db-uri", "", "Database URI")
	flag.StringVar(&secretKey, "secret-key", "", "Secret key for JWT")
	flag.StringVar(&port, "port", "50051", "Port to listen on")

}

func main() {
	flag.Parse()
	if dbURI == "" || secretKey == "" {
		fmt.Println("Please provide database URI and secret key")
		os.Exit(1)
	}
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dbURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	collection := client.Database("auth").Collection("users")

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	auth.RegisterAuthServiceServer(s, server.NewServer(collection, secretKey))
	log.Printf("Server listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
