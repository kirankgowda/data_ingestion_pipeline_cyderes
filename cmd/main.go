package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/data_ingestion_pipeline/api"

	"github.com/data_ingestion_pipeline/internal"
)


func init() {
	// Trying to load config.env file.
	err := godotenv.Load("config.env")
	if err != nil {
		log.Println("No config.env file found. Assuming Docker or system env vars are set.")
	}
}


func main() {

	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	bucket := os.Getenv("S3_BUCKET")

	log.Printf("region is %s", region)
	log.Printf("accessKey is %s", accessKey)
	log.Printf("secretKey is %s", secretKey)
	log.Printf("bucket is %s", bucket)

	if region == "" || accessKey == "" || secretKey == "" || bucket == "" {
		log.Fatal("One or more required environment variables are missing")
	}

	// Initialize S3 bucket createion if not exists once on app start
	internal.CreateS3Bucket()

	//API routes and respective handlers.
	http.HandleFunc("/ingest", api.IngestHandler)
	http.HandleFunc("/health", api.HealthHandler)
	http.HandleFunc("/getdata", api.GetDataHandler)
	http.HandleFunc("/listfiles", api.ListFilesHandler)
	http.HandleFunc("/latest", api.GetLatestIngestionHandler)
	http.HandleFunc("/delete", api.DeleteFileHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("Starting server on port %s...\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
