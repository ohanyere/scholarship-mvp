output "controller_role_arn" {
  description = "IAM role ARN for the future Karpenter controller service account."
  value       = aws_iam_role.controller.arn
}

output "karpenter_node_role_name" {
  description = "IAM role name for Karpenter-managed nodes."
  value       = aws_iam_role.node.name
}

output "karpenter_node_role_arn" {
  description = "IAM role ARN for Karpenter-managed nodes."
  value       = aws_iam_role.node.arn
}

output "karpenter_instance_profile_name" {
  description = "Instance profile name for Karpenter-managed nodes."
  value       = aws_iam_instance_profile.node.name
}
