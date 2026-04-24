locals {
  oidc_provider_hostpath = replace(var.cluster_oidc_issuer_url, "https://", "")
}

resource "aws_iam_role" "controller" {
  name = "${var.cluster_name}-karpenter-controller"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.cluster_oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${local.oidc_provider_hostpath}:aud" = "sts.amazonaws.com"
            "${local.oidc_provider_hostpath}:sub" = "system:serviceaccount:${var.service_account_namespace}:${var.service_account_name}"
          }
        }
      }
    ]
  })

  tags = var.common_tags
}

# checkov:skip=CKV_AWS_288:Karpenter needs broad read access across EC2 pricing, subnet, AMI, and capacity APIs to discover launch options dynamically.
# checkov:skip=CKV_AWS_290:Karpenter creates and terminates compute on demand, so these EC2 actions are intentionally wildcarded until workload-specific scoping is introduced.
# checkov:skip=CKV_AWS_355:This controller policy is intentionally broader in the dev platform because Karpenter provisioning and discovery are not yet partitioned by dedicated IAM conditions.
resource "aws_iam_role_policy" "controller" {
  name = "${var.cluster_name}-karpenter-controller"
  role = aws_iam_role.controller.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ec2:CreateFleet",
          "ec2:CreateLaunchTemplate",
          "ec2:CreateTags",
          "ec2:DeleteLaunchTemplate",
          "ec2:DescribeAvailabilityZones",
          "ec2:DescribeImages",
          "ec2:DescribeInstances",
          "ec2:DescribeInstanceTypeOfferings",
          "ec2:DescribeInstanceTypes",
          "ec2:DescribeLaunchTemplates",
          "ec2:DescribeSecurityGroups",
          "ec2:DescribeSpotPriceHistory",
          "ec2:DescribeSubnets",
          "ec2:DescribeTags",
          "ec2:DescribeVolumes",
          "ec2:RunInstances",
          "ec2:TerminateInstances",
          "pricing:GetProducts",
          "ssm:GetParameter",
          "eks:DescribeCluster"
        ]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = ["iam:PassRole"]
        Resource = aws_iam_role.node.arn
      }
    ]
  })
}

resource "aws_iam_role" "node" {
  name = "${var.cluster_name}-karpenter-node"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = merge(
    var.common_tags,
    {
      "karpenter.sh/discovery" = var.cluster_name
    }
  )
}

resource "aws_iam_role_policy_attachment" "node_worker" {
  role       = aws_iam_role.node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
}

resource "aws_iam_role_policy_attachment" "node_cni" {
  role       = aws_iam_role.node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
}

resource "aws_iam_role_policy_attachment" "node_ecr" {
  role       = aws_iam_role.node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_role_policy_attachment" "node_ssm" {
  role       = aws_iam_role.node.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "node" {
  name = "${var.cluster_name}-karpenter-node"
  role = aws_iam_role.node.name

  tags = var.common_tags
}
