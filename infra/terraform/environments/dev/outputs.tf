output "cluster_name" {
  description = "EKS cluster name."
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "EKS cluster endpoint."
  value       = module.eks.cluster_endpoint
}

output "cluster_certificate_authority_data" {
  description = "EKS cluster certificate authority data."
  value       = module.eks.cluster_certificate_authority_data
}

output "cluster_oidc_issuer_url" {
  description = "EKS cluster OIDC issuer URL."
  value       = module.eks.cluster_oidc_issuer_url
}

output "cluster_oidc_provider_arn" {
  description = "EKS cluster OIDC provider ARN."
  value       = module.eks.cluster_oidc_provider_arn
}

output "node_group_name" {
  description = "Baseline managed node group name."
  value       = module.eks.node_group_name
}

output "karpenter_node_role_name" {
  description = "IAM role name used by Karpenter-managed nodes."
  value       = module.karpenter.karpenter_node_role_name
}

output "karpenter_instance_profile_name" {
  description = "Instance profile used by Karpenter-managed nodes."
  value       = module.karpenter.karpenter_instance_profile_name
}

output "public_subnet_ids" {
  description = "Public subnet IDs."
  value       = module.vpc.public_subnet_ids
}

output "private_subnet_ids" {
  description = "Private subnet IDs."
  value       = module.vpc.private_subnet_ids
}

output "vpc_id" {
  description = "VPC ID."
  value       = module.vpc.vpc_id
}

output "aws_region" {
  description = "AWS region for this environment."
  value       = var.aws_region
}