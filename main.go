package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Config holds the AWS configuration
type Config struct {
	S3BucketName string
	Region       string
}

var config Config
var s3Client *s3.S3

// Initialize AWS session and S3 client
func init() {
	config = Config{
		S3BucketName: os.Getenv("AWS_S3_BUCKET"),
		Region:       os.Getenv("AWS_REGION"),
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}
	s3Client = s3.New(sess)
}

// PresignedURLResponse represents the JSON response containing the presigned URL
type PresignedURLResponse struct {
	URL string `json:"url"`
}

// GeneratePresignedURL generates a pre-signed URL for uploading to S3
func GeneratePresignedURL(w http.ResponseWriter, r *http.Request) {
	// Parse the file name from the query parameters
	fileName := r.URL.Query().Get("fileName")
	if fileName == "" {
		http.Error(w, "fileName is required", http.StatusBadRequest)
		return
	}

	// Set S3 object key and expiration
	objectKey := fmt.Sprintf("uploads/%s", fileName)
	expiration := time.Minute * 15

	// Generate the pre-signed URL
	req, _ := s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(config.S3BucketName),
		Key:    aws.String(objectKey),
	})
	presignedURL, err := req.Presign(expiration)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate pre-signed URL: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the URL as JSON
	response := PresignedURLResponse{URL: presignedURL}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck ensures the server is up and running
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// main sets up the HTTP server
func main() {
	// Check for required environment variables
	if config.S3BucketName == "" || config.Region == "" {
		log.Fatal("AWS_S3_BUCKET and AWS_REGION environment variables are required")
	}

	// Setup HTTP server routes
	mux := http.NewServeMux()
	mux.HandleFunc("/generate-presigned-url", GeneratePresignedURL)
	mux.HandleFunc("/health", HealthCheck)

	// Start the server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Println("Starting server on :8080")
	log.Fatal(server.ListenAndServe())
}
