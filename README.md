# Bedrock IAM Proxy

A secure, production-ready AWS Bedrock proxy with embedded IAM role authentication designed for EKS environments.

## Features

- **Security-First Design**: Distroless container, non-root execution, comprehensive security scanning
- **AWS Integration**: Native EKS IRSA support with fallback to EC2 instance profiles
- **Observability**: Prometheus metrics, structured logging, health checks
- **Production-Ready**: Graceful shutdowns, proper error handling, comprehensive testing
- **Private VPC Support**: Designed for fully private EKS clusters with VPC endpoints

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client        │───▶│  Bedrock Proxy  │───▶│  AWS Bedrock    │
│                 │    │  (EKS Pod)      │    │  (VPC Endpoint) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   IAM Role      │
                       │   (IRSA)        │
                       └─────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker with BuildKit
- AWS CLI configured
- kubectl access to EKS cluster

### Build

```bash
# Build the application
go build -v ./cmd/bedrock-proxy

# Run tests
go test ./...

# Build Docker image
docker build -f build/Dockerfile -t bedrock-proxy .
```

### Deploy

1. **Deploy Infrastructure**:
   ```bash
   cd deployments/terraform
   terraform init
   terraform plan
   terraform apply
   ```

2. **Deploy Application**:
   ```bash
   kubectl apply -f deployments/kubernetes/
   ```

3. **Verify Deployment**:
   ```bash
   kubectl get pods -n bedrock-system
   kubectl logs -f deployment/bedrock-proxy -n bedrock-system
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `AWS_REGION` | AWS region | `us-east-1` |
| `GIN_MODE` | Gin mode (debug/release) | `release` |
| `LOG_LEVEL` | Logging level | `info` |
| `AWS_ROLE_ARN` | IAM role ARN (auto-set by IRSA) | - |
| `AWS_WEB_IDENTITY_TOKEN_FILE` | Token file path (auto-set by IRSA) | - |

### AWS Permissions

The proxy requires the following IAM permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel",
        "bedrock:InvokeModelWithResponseStream",
        "bedrock:ListFoundationModels",
        "bedrock:GetFoundationModel"
      ],
      "Resource": "*"
    }
  ]
}
```

## API Usage

### Health Endpoints

- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

### Bedrock Proxy

- `POST /v1/bedrock/invoke-model` - Invoke Bedrock model
- `POST /bedrock/invoke-model` - Alternative endpoint

Example request:
```bash
curl -X POST http://localhost:8080/v1/bedrock/invoke-model \
  -H "Content-Type: application/json" \
  -d '{
    "modelId": "amazon.titan-text-express-v1",
    "contentType": "application/json",
    "accept": "application/json",
    "body": "{\"inputText\":\"Hello World\"}"
  }'
```

## Security

### Built-in Security Features

- **Container Security**: Distroless base image, non-root execution
- **Network Security**: Private VPC deployment, network policies
- **Authentication**: AWS IAM with IRSA integration
- **Monitoring**: Comprehensive logging and metrics
- **Compliance**: OWASP, NVD, and Trivy scanning in CI/CD

### Security Scanning

The project includes comprehensive security scanning:

```bash
# OWASP Dependency Check
dependency-check --project bedrock-proxy --scan .

# Trivy container scan
trivy image bedrock-proxy:latest

# Go security check
gosec ./...
```

## Development

### Project Structure

```
.
├── cmd/bedrock-proxy/          # Main application
├── internal/                   # Private application code
│   ├── auth/                  # AWS authentication
│   ├── health/                # Health checking
│   ├── middleware/            # HTTP middleware
│   └── proxy/                 # Bedrock proxy logic
├── pkg/                       # Public packages
│   └── metrics/              # Prometheus metrics
├── deployments/              # Deployment configurations
│   ├── kubernetes/           # K8s manifests
│   └── terraform/            # Infrastructure code
├── build/                    # Build configurations
│   └── Dockerfile           # Multi-stage Dockerfile
└── .github/workflows/        # CI/CD pipelines
```

### Adding New Features

1. Add code in appropriate `internal/` package
2. Add tests with `_test.go` suffix
3. Update metrics in `pkg/metrics/`
4. Update documentation

### Testing

```bash
# Unit tests
go test ./...

# Integration tests (requires AWS credentials)
go test ./... -tags=integration

# Benchmark tests
go test -bench=. ./...
```

## Monitoring

### Metrics

The proxy exposes Prometheus metrics at `/metrics`:

- `bedrock_proxy_requests_total` - Total requests
- `bedrock_proxy_request_duration_seconds` - Request duration
- `http_requests_total` - HTTP request count
- `health_check_status` - Health status

### Logging

Structured JSON logging with:
- Request ID correlation
- AWS credential events
- Error details
- Performance metrics

### Alerting

Example Prometheus alerting rules:

```yaml
- alert: BedrockProxyDown
  expr: up{job="bedrock-proxy"} == 0
  for: 1m
  annotations:
    summary: "Bedrock proxy is down"

- alert: BedrockProxyHighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 2m
  annotations:
    summary: "High error rate in Bedrock proxy"
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**:
   - Check IAM role permissions
   - Verify IRSA configuration
   - Check AWS credentials in logs

2. **Connection Timeouts**:
   - Verify VPC endpoints
   - Check security groups
   - Review network policies

3. **High Memory Usage**:
   - Check request patterns
   - Monitor concurrent connections
   - Review resource limits

### Debug Mode

Enable debug logging:
```bash
export GIN_MODE=debug
export LOG_LEVEL=debug
./bedrock-proxy
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

Licensed under the Apache License, Version 2.0 - see LICENSE file for details.

## Support

For issues and questions:
- Create GitHub issue
- Check troubleshooting guide
- Review logs and metrics