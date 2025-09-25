package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// AWSSigner handles AWS Signature V4 signing for Bedrock requests
type AWSSigner struct {
	credentials aws.CredentialsProvider
	region      string
	service     string
}

// NewAWSSigner creates a new AWS signer with EKS-optimized credential chain
func NewAWSSigner(region, service string) (*AWSSigner, error) {
	// Try to load config with Web Identity Token (EKS IRSA)
	cfg, err := loadAWSConfig(region)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	return &AWSSigner{
		credentials: cfg.Credentials,
		region:      region,
		service:     service,
	}, nil
}

// loadAWSConfig loads AWS configuration with EKS-optimized credential chain
func loadAWSConfig(region string) (aws.Config, error) {
	// First, try Web Identity Token (EKS IRSA)
	if roleArn := os.Getenv("AWS_ROLE_ARN"); roleArn != "" {
		if tokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE"); tokenFile != "" {
			cfg, err := config.LoadDefaultConfig(context.TODO(),
				config.WithRegion(region),
				config.WithCredentialsProvider(
					stscreds.NewWebIdentityRoleProvider(
						sts.NewFromConfig(aws.Config{Region: region}),
						roleArn,
						stscreds.IdentityTokenFile(tokenFile),
					),
				),
			)
			if err == nil {
				return cfg, nil
			}
		}
	}

	// Fallback to EC2 instance profile
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(ec2rolecreds.New()),
	)
	if err == nil {
		return cfg, nil
	}

	// Final fallback to environment variables or default credential chain
	return config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
}

// SignRequest signs an HTTP request using AWS Signature V4
func (s *AWSSigner) SignRequest(req *http.Request, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	creds, err := s.credentials.Retrieve(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve AWS credentials: %w", err)
	}

	// Validate credentials
	if creds.AccessKeyID == "" {
		return fmt.Errorf("AWS access key ID is empty")
	}
	if creds.SecretAccessKey == "" {
		return fmt.Errorf("AWS secret access key is empty")
	}

	return s.signRequestWithCredentials(req, body, creds)
}

// signRequestWithCredentials performs the actual signing
func (s *AWSSigner) signRequestWithCredentials(req *http.Request, body []byte, creds aws.Credentials) error {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	// Set required headers
	req.Header.Set("Host", req.Host)
	req.Header.Set("X-Amz-Date", amzDate)

	if creds.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", creds.SessionToken)
	}

	// Create canonical request
	canonicalRequest, signedHeaders := s.createCanonicalRequest(req, body)

	// Create string to sign
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s.service)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%x",
		algorithm,
		amzDate,
		credentialScope,
		sha256.Sum256([]byte(canonicalRequest)),
	)

	// Calculate signature
	signingKey := s.getSignatureKey(dateStamp, s.region, s.service, creds.SecretAccessKey)
	signature := hex.EncodeToString(s.hmacSHA256(signingKey, stringToSign))

	// Create authorization header
	authHeader := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		algorithm,
		creds.AccessKeyID,
		credentialScope,
		signedHeaders,
		signature,
	)

	req.Header.Set("Authorization", authHeader)
	return nil
}

// createCanonicalRequest creates the canonical request string
func (s *AWSSigner) createCanonicalRequest(req *http.Request, body []byte) (string, string) {
	// Canonical headers
	var headerNames []string
	for name := range req.Header {
		headerNames = append(headerNames, strings.ToLower(name))
	}
	sort.Strings(headerNames)

	var canonicalHeaders, signedHeaders strings.Builder
	for i, name := range headerNames {
		if i > 0 {
			signedHeaders.WriteString(";")
		}
		signedHeaders.WriteString(name)

		canonicalHeaders.WriteString(name)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(name)))
		canonicalHeaders.WriteString("\n")
	}

	// Payload hash
	payloadHash := fmt.Sprintf("%x", sha256.Sum256(body))

	// Canonical request
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders.String(),
		signedHeaders.String(),
		payloadHash,
	)

	return canonicalRequest, signedHeaders.String()
}

// getSignatureKey generates the signing key
func (s *AWSSigner) getSignatureKey(dateStamp, region, service, secretKey string) []byte {
	kDate := s.hmacSHA256([]byte("AWS4"+secretKey), dateStamp)
	kRegion := s.hmacSHA256(kDate, region)
	kService := s.hmacSHA256(kRegion, service)
	kSigning := s.hmacSHA256(kService, "aws4_request")
	return kSigning
}

// hmacSHA256 creates HMAC-SHA256 hash
func (s *AWSSigner) hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}