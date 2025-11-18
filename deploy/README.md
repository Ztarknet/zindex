# Zindex Deployment Guide

This directory contains the Kubernetes/Helm deployment configuration for the Zindex service.

## Prerequisites

- GCP account with GKE Autopilot cluster
- `gcloud` CLI configured
- `kubectl` configured to access your cluster
- `helm` v3 installed
- Docker configured and logged in to your registry

## Architecture

The deployment consists of two main services:
1. **PostgreSQL Database** - Persistent storage for indexed data
2. **Zindex Server** - Go indexer service

## Initial Setup

### 1. Create GKE Autopilot Cluster

```bash
gcloud container clusters create-auto zindex \
  --region=us-central1 \
  --project=YOUR_PROJECT_ID
```

### 2. Configure kubectl

```bash
gcloud container clusters get-credentials zindex \
  --region=us-central1 \
  --project=YOUR_PROJECT_ID
```

### 3. Reserve Static IP

```bash
gcloud compute addresses create zindex-static-ip \
  --global \
  --ip-version IPV4 \
  --project=YOUR_PROJECT_ID
```

Get the IP address:
```bash
gcloud compute addresses describe zindex-static-ip \
  --global \
  --format="get(address)"
```

### 4. Configure DNS

Point your domain (e.g., `zindex.ztarknet.cash`) to the static IP address from step 3.

### 5. Update Managed Certificate

Edit `deploy/managed-cert.yaml` and update the domain:

```yaml
spec:
  domains:
  - zindex.ztarknet.cash  # Your actual domain
```

### 6. Deploy Managed Certificate

```bash
kubectl apply -f deploy/managed-cert.yaml
```

**Note:** It may take up to 15-20 minutes for the certificate to provision. You can check status with:

```bash
kubectl describe managedcertificate zindex-managed-cert
```

## Deployment

### Using Make Commands (Recommended)

```bash
# Build and push Docker image
make docker-build-prod
make docker-push

# Deploy/upgrade Helm chart
make helm-install
# or
make helm-upgrade
```

### Manual Deployment

#### 1. Build Production Docker Image

```bash
# Build for linux/amd64 platform
docker build --platform linux/amd64 -f Dockerfile.prod -t brandonjroberts/zindex:latest .

# Tag with version
export APP_VERSION="v0.1.0"
export COMMIT_SHA=$(git rev-parse --short HEAD)
docker tag brandonjroberts/zindex:latest brandonjroberts/zindex:${APP_VERSION}-${COMMIT_SHA}

# Push to registry
docker push brandonjroberts/zindex:latest
docker push brandonjroberts/zindex:${APP_VERSION}-${COMMIT_SHA}
```

#### 2. Configure Values

Edit `deploy/zindex-infra/values.yaml` and update:

- `postgres.password` - Set a strong password
- `zindex.rpc_url` - Update to your production RPC endpoint
- `deployments.zindex.image` - Your Docker registry
- `deployments.zindex.tag` - Your image tag

#### 3. Install with Helm

```bash
# Install
helm install zindex deploy/zindex-infra \
  --set postgres.password=YOUR_STRONG_PASSWORD \
  --set deployments.zindex.commit_sha=${COMMIT_SHA}

# Or upgrade existing deployment
helm upgrade zindex deploy/zindex-infra \
  --set postgres.password=YOUR_STRONG_PASSWORD \
  --set deployments.zindex.commit_sha=${COMMIT_SHA}
```

## Verification

### Check Deployment Status

```bash
# Check all resources
kubectl get all

# Check pods
kubectl get pods

# Check services
kubectl get services

# Check ingress
kubectl get ingress

# Check managed certificate status
kubectl describe managedcertificate zindex-managed-cert
```

### View Logs

```bash
# Zindex server logs
kubectl logs -f deployment/zindex-server

# PostgreSQL logs
kubectl logs -f deployment/zindex-postgres
```

### Test API

```bash
# Once certificate is provisioned (may take 15-20 minutes)
curl https://zindex.ztarknet.cash/health
```

## Updating the Deployment

### Code Changes

```bash
# 1. Build new image with commit SHA
export COMMIT_SHA=$(git rev-parse --short HEAD)
make docker-build-prod
make docker-push

# 2. Upgrade Helm deployment
make helm-upgrade
```

### Configuration Changes

```bash
# Edit values.yaml
vim deploy/zindex-infra/values.yaml

# Apply changes
helm upgrade zindex deploy/zindex-infra
```

## Rollback

```bash
# View release history
helm history zindex

# Rollback to previous version
helm rollback zindex

# Rollback to specific revision
helm rollback zindex 2
```

## Uninstall

```bash
# Uninstall Helm release
helm uninstall zindex

# Delete managed certificate
kubectl delete -f deploy/managed-cert.yaml

# Delete persistent volumes (optional - this will DELETE all data)
kubectl delete pvc zindex-postgres-volume-claim
```

## Monitoring

### Resource Usage

```bash
# Check pod resource usage
kubectl top pods

# Check node resource usage
kubectl top nodes
```

### Database Access

```bash
# Port forward to PostgreSQL
kubectl port-forward service/zindex-postgres 5432:5432

# Connect with psql
psql -h localhost -U zindex -d zindex
```

## Troubleshooting

### Pod Not Starting

```bash
# Describe pod for events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>

# Get previous logs if pod restarted
kubectl logs <pod-name> --previous
```

### Certificate Not Provisioning

- Ensure DNS is properly configured and pointing to the static IP
- Verify the domain in `managed-cert.yaml` matches your DNS record
- Check certificate status: `kubectl describe managedcertificate zindex-managed-cert`
- GKE managed certificates can take 15-20 minutes to provision

### Database Connection Issues

- Verify PostgreSQL pod is running: `kubectl get pods`
- Check PostgreSQL logs: `kubectl logs deployment/zindex-postgres`
- Verify service is exposed: `kubectl get service zindex-postgres`

## Cost Optimization

- Start with minimal resources (current config uses 1 replica for each service)
- Monitor usage with `kubectl top` commands
- Adjust replica counts in `values.yaml` as needed
- Consider using Cloud SQL for production databases (managed, backups included)

## Security Notes

- Change the default PostgreSQL password in `values.yaml`
- Use Kubernetes secrets for sensitive data in production
- Consider using GCP Secret Manager with External Secrets Operator
- Restrict CORS origins in production config
- Disable admin endpoints in production (`zindex.admin: false`)

## Next Steps

- Set up monitoring with Prometheus/Grafana
- Configure log aggregation
- Set up automated backups for PostgreSQL
- Implement CI/CD pipeline for automated deployments
- Add health checks and readiness probes
