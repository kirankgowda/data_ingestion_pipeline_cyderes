package internal

import (
	"time"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"log"
	
	"github.com/google/uuid"
)

// used to Transform the ingested data by adding ingested_at and source fields to all the data elements
func transformData(data []map[string]interface{}) ([]map[string]interface{},error) {

	if data == nil {
		return nil, errors.New("input data is nil")
	}

	now := time.Now().UTC().Format(time.RFC3339)
	for _, item := range data {
		item["ingested_at"] = now
		item["source"] = "placeholder_api"
	}
	return data, nil
}

//Struct that represents IngestResult metadata response object
type IngestResult struct {
	Filename string `json:"filename"`
	Count    int    `json:"count"`
}

// IngestAndStore fetches, transforms, and uploads data to S3 bucket; returns metadata or error if any
func IngestAndStore() (*IngestResult,error) {

	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		return nil, errors.New("failed to fetch external data: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil,errors.New("failed to fetch external data: status code " + http.StatusText(resp.StatusCode))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil,errors.New("failed to read response body: " + err.Error())
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, errors.New("invalid response format: " + err.Error())
	}

	transformed, err := transformData(data)
	if err != nil {
		return nil, errors.New("Error during Response transformation: " + err.Error())
	}

	id := uuid.New()
	filename := "data_" + id.String() + ".json"
	log.Printf("filename  data is %s", filename)

	err = UploadToS3(filename, transformed)
	if err != nil {
		return nil,errors.New("failed to upload to S3: " + err.Error())
	}

	result := &IngestResult{
		Filename: filename,
		Count:    len(transformed),
	}

	return result, nil
}
