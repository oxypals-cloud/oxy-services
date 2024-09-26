package server

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/oxypals-cloud/oxy-services/auth/authpb" // Replace with your path
)

type server struct {
	collection *mongo.Collection
	secretKey  []byte
}

func NewServer(collection *mongo.Collection, secretKey string) *server {
	return &server{
		collection: collection,
		secretKey:  []byte(secretKey),
	}
}

func (*server) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	// Check if username already exists
	filter := bson.M{"username": req.Username}
	count, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unable to connect to database: %v", err))
	}

	if count > 0 {
		return nil, status.Errorf(codes.AlreadyExists, fmt.Sprintf("Username already exists: %v", err))
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Failed to hash password: %v", err))
	}

	// Insert user data into MongoDB
	newUser := User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
	}
	insertResult, err := s.collection.InsertOne(ctx, newUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Failed to insert user: %v", err))
	}

	log.Printf("User %s was created with ID %s", newUser.Username, insertResult.InsertedID)

	return &authpb.RegisterResponse{
		Message: "User registered successfully",
	}, nil
}

func (*server) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	// Find user by username
	filter := bson.M{"username": req.Username}
	var user User
	err := s.collection.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("No user found with this username: %v", err))
		}
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unable to connect to database: %v", err))
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Incorrect password: %v", err))
	}

	// Generate JWT token
	token, err := generateJWT(user.Username, s.secretKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Failed to generate JWT: %v", err))
	}

	return &authpb.LoginResponse{
		Token:   token,
		Message: "Login successful",
	}, nil
}

func (*server) VerifyToken(ctx context.Context, req *authpb.VerifyTokenRequest) (*authpb.VerifyTokenResponse, error) {
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, status.Errorf(codes.Unauthenticated, fmt.Sprintf("Invalid token: %v", err))
		}
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unable to parse token: %v", err))
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, fmt.Sprintf("Invalid token: %v", err))
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, fmt.Sprintf("Could not parse claims: %v", err))
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("No username found in token: %v", err))
	}

	return &authpb.VerifyTokenResponse{
		Valid:    true,
		Message:  "Token is valid",
		Username: username,
	}, nil
}

func generateJWT(username string, secretKey []byte) (string, error) {
	// Create token claims
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with your secret key
	ss, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Username string             `bson:"username"`
	Password string             `bson:"password"`
	Email    string             `bson:"email"`
}
