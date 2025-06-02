package api

import (
	"encoding/json"
	"net/http"
	"log"

	"github.com/data_ingestion_pipeline/internal"
)

// JSON response structure
type jsonResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Filename string `json:"filename,omitempty"`
	Count    int    `json:"count,omitempty"`
}


// HealthHandler is used to see if server is up and running.
// allows only GET and returns JSON
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	resp := jsonResponse{
		Status:  "success",
		Message: "API is up and running",
	}

	writeJSON(w, http.StatusOK, resp)
}


// IngestHandler is used to see to Ingest new Data from the API call, transform it and Store it in AWS S3 bucket.
// allows only POST and returns JSON.
func IngestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	result, err := internal.IngestAndStore()
	if err != nil {
		log.Printf("Ingestion error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := jsonResponse{
		Status:   "success",
		Message:  "Data ingested and uploaded to S3 successfully",
		Filename: result.Filename,
		Count:    result.Count,
	}

	writeJSON(w, http.StatusOK, resp)
}


// GetDataHandler is used to see to get the ingested File from AWS S3 bucket based on Filename
// allows only GET and returns JSON.
func GetDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing 'filename' query parameter")
		return
	}

	log.Printf("Fetching data for filename: %s", filename)

	data, err := internal.ReadJSONFromS3("logs/" + filename)
	if err != nil {
		if err.Error() == "no_such_key" {
			writeJSONError(w, http.StatusNotFound, "File not found in S3")
			return
		}
		log.Printf("Error fetching data from S3: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to fetch data from S3")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Data Fetched Successfully",
		"records": data,
	})
}


// ListFilesHandler is used to see to get the list of All Files in AWS S3 bucket.
// allows only GET and returns JSON.
func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	files, err := internal.ListFilesFromS3("logs/")
	if err != nil {
		log.Printf("Error listing files from S3: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to list files from S3")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "success",
		"message":  "Files listed successfully",
		"filenames": files,
	})
}


// GetLatestIngestionHandler returns the json data along with metadata of the latest successful ingestion
// allows only GET and returns JSON.
func GetLatestIngestionHandler(w http.ResponseWriter, r *http.Request) {
	
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	rawMeta, err := internal.ReadJSONFromS3("logs/latest_ingestion.json")
	log.Printf("metadata is %s", rawMeta)
	if err != nil {
		if err.Error() == "no_such_key" {
			writeJSONError(w, http.StatusNotFound, "File not found in S3")
			return
		}

		writeJSONError(w, http.StatusInternalServerError, "Failed to fetch latest ingestion metadata")
		return
	}

	metadata, ok := rawMeta.(map[string]interface{})
	if !ok {

		writeJSONError(w, http.StatusInternalServerError, "FInvalid metadata format")
		return
	}

	// Extract the filename and get the needed data
	filename, ok := metadata["filename"].(string)
	if !ok || filename == "" {

		writeJSONError(w, http.StatusInternalServerError, "Filename not found in metadata")
		return
	}

	dataObj, err := internal.ReadJSONFromS3("logs/" + filename)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to read ingested data")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Fetched Latest ingestion successfully",
		"metadata": metadata,
		"data":     dataObj,
	})
}


// DeleteFileHandler used to delete the File in AWS S3 based on filename 
// allows only GET and returns JSON.
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	filename := r.URL.Query().Get("filename")
	if filename == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing filename query param")
		return
	}

	err := internal.DeleteFileFromS3(filename)
	if err != nil {
		log.Printf("issue is %s", err)
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete file")
		return

	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "File deleted successfully",
	})
}


// Helper: write JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Helper: write JSON error response
func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, jsonResponse{
		Status:  "error",
		Message: message,
	})
}
