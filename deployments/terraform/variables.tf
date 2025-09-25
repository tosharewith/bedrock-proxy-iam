variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "private_subnets" {
  description = "Private subnet CIDR blocks"
  type        = list(string)
  default = [
    "10.0.1.0/24",
    "10.0.2.0/24",
    "10.0.3.0/24"
  ]
}

variable "database_subnets" {
  description = "Database subnet CIDR blocks"
  type        = list(string)
  default = [
    "10.0.21.0/24",
    "10.0.22.0/24",
    "10.0.23.0/24"
  ]
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.28"
}

variable "node_groups" {
  description = "EKS node group configurations"
  type        = any
  default = {
    bedrock_proxy = {
      name = "bedrock-proxy-nodes"
      instance_types = ["m5.large"]
      ami_type       = "AL2_x86_64"
      min_size       = 2
      max_size       = 10
      desired_size   = 3
      disk_size      = 50
      disk_type      = "gp3"

      taints = [
        {
          key    = "bedrock-proxy"
          value  = "dedicated"
          effect = "NO_SCHEDULE"
        }
      ]

      labels = {
        workload-type = "bedrock-proxy"
      }
    }

    system = {
      name = "system-nodes"
      instance_types = ["t3.medium"]
      ami_type       = "AL2_x86_64"
      min_size       = 2
      max_size       = 4
      desired_size   = 2
      disk_size      = 30
      disk_type      = "gp3"

      taints = [
        {
          key    = "CriticalAddonsOnly"
          value  = "true"
          effect = "NO_SCHEDULE"
        }
      ]
    }
  }
}