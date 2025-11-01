package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
)

// AWSSigner handles AWS Signature V4 signing for Bedrock requests
type AWSSigner struct {
	region  string
	service string
}

// NewAWSSigner creates a new AWS signer with EKS-optimized credential chain
func NewAWSSigner(region, service string) (*AWSSigner, error) {
	return &AWSSigner{
		region:  region,
		service: service,
	}, nil
}

// SignRequest signs an HTTP request using AWS Signature V4
func (s *AWSSigner) SignRequest(req *http.Request, body []byte) error {
	// Load AWS config with default credential chain (supports IRSA, EC2 instance profile, env vars)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("Unable to load AWS config: %v", err)
		return fmt.Errorf("unable to load AWS config: %w", err)
	}

	credentials, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		log.Printf("Unable to retrieve AWS credentials: %v", err)
		return fmt.Errorf("unable to retrieve AWS credentials: %w", err)
	}

	// Remove unwanted headers that might interfere with signing
	req.Header.Del("User-Agent")
	req.Header.Del("Authorization")
	req.Header.Del("Connection")
	req.Header.Del("X-Amz-Content-Sha256")
	req.Header.Del("X-Amz-Date")

	// Calculate payload hash
	payloadHash := sha256.Sum256(body)
	hash := hex.EncodeToString(payloadHash[:])

	// Use AWS SDK v4 signer
	signer := v4.NewSigner()
	err = signer.SignHTTP(context.TODO(), credentials, req, hash, s.service, s.region, time.Now().UTC())
	if err != nil {
		log.Printf("Unable to sign request: %v", err)
		return fmt.Errorf("unable to sign request: %w", err)
	}

	return nil
}
