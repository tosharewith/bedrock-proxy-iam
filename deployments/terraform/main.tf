# Complete EKS Private VPC Setup for Bedrock Proxy
# This configuration ensures all traffic stays within AWS backbone

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  required_version = ">= 1.0"
}

locals {
  cluster_name = "bedrock-proxy-cluster"
  region      = var.aws_region

  tags = {
    Environment = var.environment
    Project     = "bedrock-proxy"
    Owner       = "platform-team"
    ManagedBy   = "terraform"
  }
}

# ============================================================================
# VPC Configuration - Fully Private
# ============================================================================
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${local.cluster_name}-vpc"
  cidr = var.vpc_cidr

  azs = data.aws_availability_zones.available.names

  # Private subnets only - no public subnets for maximum security
  private_subnets = var.private_subnets

  # Isolated subnets for databases/sensitive workloads
  database_subnets = var.database_subnets

  # VPC Flow Logs for security monitoring
  enable_flow_log = true
  create_flow_log_cloudwatch_log_group = true
  create_flow_log_cloudwatch_iam_role  = true
  flow_log_destination_type = "cloud-watch-logs"

  # DNS settings for private resolution
  enable_dns_hostnames = true
  enable_dns_support   = true

  # No NAT Gateway - fully private
  enable_nat_gateway = false
  enable_vpn_gateway = false

  tags = local.tags
}

data "aws_availability_zones" "available" {
  state = "available"
}

# ============================================================================
# VPC Endpoints - Critical for Private VPC Operations
# ============================================================================

# S3 Gateway Endpoint (for container images, logs)
resource "aws_vpc_endpoint" "s3" {
  vpc_id            = module.vpc.vpc_id
  service_name      = "com.amazonaws.${local.region}.s3"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = module.vpc.private_route_table_ids

  tags = merge(local.tags, {
    Name = "${local.cluster_name}-s3-endpoint"
  })
}

# ECR API Endpoint
resource "aws_vpc_endpoint" "ecr_api" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${local.region}.ecr.api"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  security_group_ids  = [aws_security_group.vpc_endpoints.id]
  private_dns_enabled = true

  tags = merge(local.tags, {
    Name = "${local.cluster_name}-ecr-api-endpoint"
  })
}

# ECR Docker Registry Endpoint
resource "aws_vpc_endpoint" "ecr_dkr" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${local.region}.ecr.dkr"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  security_group_ids  = [aws_security_group.vpc_endpoints.id]
  private_dns_enabled = true

  tags = merge(local.tags, {
    Name = "${local.cluster_name}-ecr-dkr-endpoint"
  })
}

# ⭐ CRITICAL: Bedrock Runtime Endpoint
resource "aws_vpc_endpoint" "bedrock_runtime" {
  vpc_id              = module.vpc.vpc_id
  service_name        = "com.amazonaws.${local.region}.bedrock-runtime"
  vpc_endpoint_type   = "Interface"
  subnet_ids          = module.vpc.private_subnets
  security_group_ids  = [aws_security_group.bedrock_endpoint.id]
  private_dns_enabled = true

  # Bedrock-specific policy
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = "*"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream"
        ]
        Resource = "*"
      }
    ]
  })

  tags = merge(local.tags, {
    Name = "${local.cluster_name}-bedrock-runtime-endpoint"
  })
}

# ============================================================================
# Security Groups
# ============================================================================

# Security Group for VPC Endpoints
resource "aws_security_group" "vpc_endpoints" {
  name_prefix = "${local.cluster_name}-vpc-endpoints"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "HTTPS from VPC"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [module.vpc.vpc_cidr_block]
  }

  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.tags, {
    Name = "${local.cluster_name}-vpc-endpoints-sg"
  })
}

# Dedicated Security Group for Bedrock Endpoint
resource "aws_security_group" "bedrock_endpoint" {
  name_prefix = "${local.cluster_name}-bedrock-endpoint"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "HTTPS from EKS cluster"
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [module.eks.node_security_group_id]
  }

  ingress {
    description = "HTTPS from Bedrock proxy pods"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [module.vpc.vpc_cidr_block]
  }

  tags = merge(local.tags, {
    Name = "${local.cluster_name}-bedrock-endpoint-sg"
  })
}

# ============================================================================
# EKS Cluster - Private Configuration
# ============================================================================
module "eks" {
  source = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = local.cluster_name
  cluster_version = var.kubernetes_version

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  # Private cluster configuration
  cluster_endpoint_private_access      = true
  cluster_endpoint_public_access      = false  # Fully private
  cluster_endpoint_public_access_cidrs = []

  # Cluster encryption
  cluster_encryption_config = [
    {
      provider_key_arn = aws_kms_key.eks.arn
      resources        = ["secrets"]
    }
  ]

  # EKS Addons
  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
      configuration_values = jsonencode({
        env = {
          ENABLE_PREFIX_DELEGATION = "true"
          WARM_PREFIX_TARGET       = "1"
        }
      })
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }

  # Node groups with private configuration
  eks_managed_node_groups = var.node_groups

  tags = local.tags
}

# ============================================================================
# KMS Key for EKS Encryption
# ============================================================================
resource "aws_kms_key" "eks" {
  description             = "EKS Cluster Encryption Key"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = local.tags
}

resource "aws_kms_alias" "eks" {
  name          = "alias/${local.cluster_name}-eks"
  target_key_id = aws_kms_key.eks.key_id
}

# ============================================================================
# IAM Role for Bedrock Access (IRSA)
# ============================================================================
module "bedrock_proxy_irsa" {
  source = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"

  role_name = "${local.cluster_name}-bedrock-proxy-role"

  role_policy_arns = {
    bedrock = aws_iam_policy.bedrock_proxy.arn
  }

  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["bedrock-system:bedrock-proxy-sa"]
    }
  }

  tags = local.tags
}

# Custom IAM Policy for Bedrock Proxy
resource "aws_iam_policy" "bedrock_proxy" {
  name        = "${local.cluster_name}-bedrock-proxy-policy"
  description = "Bedrock proxy permissions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:InvokeModelWithResponseStream",
          "bedrock:ListFoundationModels",
          "bedrock:GetFoundationModel"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${local.region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/bedrock-proxy/*"
      }
    ]
  })
}

data "aws_caller_identity" "current" {}