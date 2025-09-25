output "cluster_name" {
  description = "Name of the EKS cluster"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "Base64 encoded certificate data for the cluster"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = module.vpc.vpc_id
}

output "private_subnets" {
  description = "List of private subnet IDs"
  value       = module.vpc.private_subnets
}

output "bedrock_endpoint_id" {
  description = "ID of the Bedrock VPC endpoint"
  value       = aws_vpc_endpoint.bedrock_runtime.id
}

output "bedrock_proxy_role_arn" {
  description = "ARN of the Bedrock proxy IAM role"
  value       = module.bedrock_proxy_irsa.iam_role_arn
}

output "bedrock_proxy_role_name" {
  description = "Name of the Bedrock proxy IAM role"
  value       = module.bedrock_proxy_irsa.iam_role_name
}

output "kms_key_id" {
  description = "KMS key ID for EKS encryption"
  value       = aws_kms_key.eks.id
}

output "kms_key_arn" {
  description = "KMS key ARN for EKS encryption"
  value       = aws_kms_key.eks.arn
}