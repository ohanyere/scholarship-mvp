output "cluster_name" {
  description = "EKS cluster name."
  value       = aws_eks_cluster.this.name
}

output "cluster_endpoint" {
  description = "EKS cluster API endpoint."
  value       = aws_eks_cluster.this.endpoint
}

output "cluster_certificate_authority_data" {
  description = "Base64 certificate authority data for the cluster."
  value       = aws_eks_cluster.this.certificate_authority[0].data
}

output "cluster_oidc_issuer_url" {
  description = "OIDC issuer URL for the cluster."
  value       = aws_eks_cluster.this.identity[0].oidc[0].issuer
}

output "cluster_oidc_provider_arn" {
  description = "OIDC provider ARN for IAM roles for service accounts."
  value       = aws_iam_openid_connect_provider.this.arn
}

output "node_group_name" {
  description = "Baseline managed node group name."
  value       = aws_eks_node_group.baseline.node_group_name
}

output "node_group_role_name" {
  description = "Baseline managed node IAM role name."
  value       = aws_iam_role.node_group.name
}
