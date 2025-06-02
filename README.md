# Data Ingestion Service

This is a Go-based microservice for ingesting data from an external API, transforming it, and storing the result in AWS S3. It exposes RESTful endpoints for triggering ingestion, viewing stored data, and checking health status.

## Features

- Fetches logs/data from an external API
- Transforms data with metadata
- Stores transformed data as JSON in AWS S3
- Provides REST API with Swagger documentation
- Deployable to AWS ECS using Docker
- Configurable via `config.env` file

---

## Prerequisites

- AWS Account
- IAM User with:
  - **AmazonS3FullAccess**
  - **AmazonECS_FullAccess**
- Docker and Docker Compose installed
- Go 1.22+ installed (for local dev and testing)

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/kirankgowda/data_ingestion_pipeline_cyderes.git
cd data_ingestion_pipeline_cyderes
```

--- 

## Running the Application.

### Configure Environment Variables

Edit the config.env file and add your AWS credentials and S3 bucket name:

```bash
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
AWS_REGION=your-region
S3_BUCKET_NAME=your-s3-bucket
PORT=PORT(5000)
```

### Run the Service with Docker Compose

```bash
docker-compose up --build
```

Access the Application with http://localhost:5000/health

### Run the Application locally (with Go 1.22 installed).

Check if you are in the application folder (data_ingestion_pipeline_cyderes)

Run below command to bring up the application

```bash 
go mod tidy
go run ./cmd/main.go
```

### Run Unit Tests

Run Basic Unit Tests once you are in the Container's folder structure.

```bash
go test ./... -v
```

---

## API Documentation

Refer to *data_ingestion_api.yaml* file in the project folder for the API usage details.

Navigate to https://editor.swagger.io/ site, Copy and Paste the *data_ingestion_api.yaml* contents in the site to see the API Documentation.


--- 

## Deploying to AWS ECS

### 1. Create a AWS Cluster

Open AWS Console → ECS → Create Cluster → Select EC2 or Fargate

--- 

### 2. Create Task Definition

Register a new task definition

Use Docker image from Docker Hub (or ECR if applicable)

Define container port (5000) and environment variables

--- 

### 3. Create a Service

Choose your ECS Cluster

Launch the task definition

Set desired number of tasks (replicas)

Choose networking (VPC, subnets, security group)

--- 

### 4. Launch the Service

The container will now run on AWS ECS

The Tasks will be sceduled based on the replicas set before.

---

## GitHub Actions CI/CD

To configure GitHub Actions for automatic builds and ECS deployment:

Set the following repository secrets:

```bash
AWS_ACCESS_KEY_ID
AWS_REGION
AWS_SECRET_ACCESS_KEY
CONTAINER_NAME
DOCKERHUB_TOKEN
DOCKERHUB_USERNAME
ECS_CLUSTER
ECS_SERVICE
ECS_TASK_FAMILY
IMAGE_NAME
PORT
```

Workflow Trigger is available in the path - .github/workflows/deploy.yml

The Workflow will perform below Steps to Build and Push the Application In AWS ECS Instance :--

 - Refers to the codebase Main Branch of this project
 - Initiates CI/CD steps once a checkin happens to the main branch or manually.
 - Setups the GO project, builds and run the tests.
 - Builds the Docker Image.
 - Logs In to Docker hub and pushes the latest image to the provided docker hub image.
 - Logs In to AWS and creates a new Task definition using the recently pushed Image.
 - Launches/updates the Service with a new TASK Instance of the application. 







