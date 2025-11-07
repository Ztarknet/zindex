# Deploying zIndex to GCP

This guide provides instructions for deploying the zIndex Zcash indexer to Google Cloud Platform.

## Prerequisites

- Google Cloud SDK installed and configured
- Docker installed locally
- A GCP project with billing enabled
- Appropriate IAM permissions for the following services:
  - Cloud Run or Compute Engine
  - Cloud SQL (PostgreSQL)
  - Container Registry or Artifact Registry
  - VPC Network (for private connectivity)

## Deployment Options

### Option 1: Cloud Run (Recommended for Serverless)

Cloud Run is recommended for automatic scaling and serverless deployment.

#### 1. Set up Cloud SQL PostgreSQL

```bash
# Create a PostgreSQL instance
gcloud sql instances create zindex-db \
  --database-version=POSTGRES_15 \
  --tier=db-custom-2-8192 \
  --region=us-central1

# Create the database
gcloud sql databases create zindex --instance=zindex-db

# Create a user
gcloud sql users create zindex \
  --instance=zindex-db \
  --password=SECURE_PASSWORD_HERE
```

#### 2. Build and Push Docker Image

```bash
# Set your project ID
export PROJECT_ID=your-project-id
export REGION=us-central1

# Build the Docker image
docker build -t gcr.io/${PROJECT_ID}/zindex:latest .

# Push to Google Container Registry
docker push gcr.io/${PROJECT_ID}/zindex:latest
```

#### 3. Deploy to Cloud Run

```bash
# Deploy with Cloud SQL connection
gcloud run deploy zindex \
  --image gcr.io/${PROJECT_ID}/zindex:latest \
  --region ${REGION} \
  --platform managed \
  --allow-unauthenticated \
  --add-cloudsql-instances ${PROJECT_ID}:${REGION}:zindex-db \
  --set-env-vars "CONFIG_PATH=/root/configs/config.yaml" \
  --memory 2Gi \
  --cpu 2 \
  --timeout 3600 \
  --max-instances 10
```

### Option 2: Compute Engine (For Long-Running Indexer)

For a continuously running indexer, Compute Engine provides more control.

#### 1. Create a VM Instance

```bash
# Create a VM with container-optimized OS
gcloud compute instances create-with-container zindex \
  --zone=us-central1-a \
  --machine-type=e2-medium \
  --container-image=gcr.io/${PROJECT_ID}/zindex:latest \
  --container-restart-policy=always \
  --boot-disk-size=50GB \
  --tags=zindex-server
```

#### 2. Set up Firewall Rules

```bash
# Allow API traffic
gcloud compute firewall-rules create allow-zindex-api \
  --allow tcp:8080 \
  --target-tags zindex-server \
  --source-ranges 0.0.0.0/0
```

#### 3. SSH and Configure

```bash
# SSH into the instance
gcloud compute ssh zindex --zone=us-central1-a

# View logs
docker logs $(docker ps -q)
```

### Option 3: GKE (Kubernetes for Production)

For production deployments with high availability and complex orchestration needs.

#### 1. Create GKE Cluster

```bash
gcloud container clusters create zindex-cluster \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type e2-standard-2 \
  --enable-autoscaling \
  --min-nodes 1 \
  --max-nodes 5
```

#### 2. Apply Kubernetes Manifests

See `kubernetes/` directory for deployment, service, and configmap manifests.

```bash
kubectl apply -f deploy/kubernetes/
```

## Configuration

### Environment Variables

Update the `configs/config.yaml` file or override with environment variables:

- `RPC_URL`: Zcash RPC endpoint
- `RPC_USERNAME`: RPC username
- `RPC_PASSWORD`: RPC password
- `DB_HOST`: PostgreSQL host (use Cloud SQL proxy for Cloud Run)
- `DB_PORT`: PostgreSQL port
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name

### Cloud SQL Proxy (for Cloud Run)

When using Cloud Run with Cloud SQL, use the Unix socket connection:

```yaml
database:
  host: "/cloudsql/PROJECT_ID:REGION:INSTANCE_NAME"
  port: "5432"
```

## Monitoring and Logging

### View Logs

```bash
# Cloud Run logs
gcloud run logs read zindex --region=${REGION}

# Compute Engine logs
gcloud compute instances get-serial-port-output zindex --zone=us-central1-a
```

### Set Up Alerts

Configure Cloud Monitoring alerts for:
- High error rates
- Memory usage > 80%
- CPU usage > 80%
- Database connection failures

## Scaling

### Cloud Run Auto-scaling

Cloud Run automatically scales based on incoming requests. Configure:

```bash
gcloud run services update zindex \
  --min-instances 1 \
  --max-instances 10 \
  --concurrency 80
```

### Compute Engine Auto-scaling

For Compute Engine, use managed instance groups:

```bash
gcloud compute instance-groups managed create zindex-group \
  --base-instance-name zindex \
  --size 1 \
  --template zindex-template \
  --zone us-central1-a

gcloud compute instance-groups managed set-autoscaling zindex-group \
  --max-num-replicas 5 \
  --min-num-replicas 1 \
  --target-cpu-utilization 0.75 \
  --zone us-central1-a
```

## Security Best Practices

1. Use Secret Manager for sensitive credentials
2. Enable VPC Service Controls
3. Use private IP addresses for database connections
4. Enable Cloud Armor for DDoS protection
5. Implement proper IAM roles and permissions
6. Enable audit logging

## Cost Optimization

- Use preemptible VMs for non-critical workloads
- Set up committed use discounts for predictable workloads
- Monitor with Cloud Billing budgets and alerts
- Use Cloud SQL read replicas only when needed
- Implement proper resource cleanup policies

## Troubleshooting

### Common Issues

1. **Database connection failures**: Check Cloud SQL proxy configuration and network settings
2. **Out of memory errors**: Increase memory allocation in Cloud Run or VM instance
3. **Slow indexing**: Increase batch size or add more instances
4. **API timeouts**: Increase timeout settings and check RPC endpoint health

### Health Checks

Add health check endpoints to your application and configure:

```bash
gcloud run services update zindex \
  --region=${REGION} \
  --startup-cpu-boost \
  --cpu-throttling
```
