resource "aws_iam_role" "k8s-acc-cluster" {
  name = local.cluster_name

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "k8s-acc-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.k8s-acc-cluster.name
}

resource "aws_iam_role_policy_attachment" "k8s-acc-AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.k8s-acc-cluster.name
}

resource "aws_iam_role" "k8s-acc-node" {
  name = "${local.cluster_name}-node"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "k8s-acc-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.k8s-acc-node.name
}

resource "aws_iam_role_policy_attachment" "k8s-acc-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.k8s-acc-node.name
}

resource "aws_iam_role_policy_attachment" "k8s-acc-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.k8s-acc-node.name
}

# Policies to allow nodes to access S3 buckets with Cortex storage

resource "aws_iam_policy" "cortex-operator-s3-policy" {
  name        = "cortex-operator-s3-policy"
  description = "S3 policy to access buckets with Cortex block storage data and alert manager configuration"

  policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [
      {
        "Sid": "ListObjectsInBucket",
        "Effect": "Allow",
        "Action": ["s3:ListBucket"],
        "Resource": [
          aws_s3_bucket.cortex_data_bucket.arn , "${aws_s3_bucket.cortex_data_bucket.arn}/*",
          aws_s3_bucket.cortex_config_bucket.arn , "${aws_s3_bucket.cortex_config_bucket.arn}/*",
        ]
      },
      {
        "Sid": "AllObjectActions",
        "Effect": "Allow",
        "Action": "s3:*Object",
        "Resource": [
          aws_s3_bucket.cortex_data_bucket.arn , "${aws_s3_bucket.cortex_data_bucket.arn}/*",
          aws_s3_bucket.cortex_config_bucket.arn , "${aws_s3_bucket.cortex_config_bucket.arn}/*",
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "cortex-operator-s3-attach" {
  role       = aws_iam_role.k8s-acc-node.name
  policy_arn = aws_iam_policy.cortex-operator-s3-policy.arn
}

# resource "aws_iam_policy" "cortex-operator-storage-config-s3" {
#   name        = "cortex-operator-storage-config-s3"
#   description = "S3 policy to manage bucket with Cortex AlertManager and Rule configuration"

#   policy = jsonencode({
#     "Version": "2012-10-17",
#     "Statement": [
#       {
#         "Sid": "ListObjectsInBucket",
#         "Effect": "Allow",
#         "Action": ["s3:ListBucket"],
#         "Resource": [
#           aws_s3_bucket ,
#           "${aws_s3_bucket }/*"
#         ]
#       },
#       {
#         "Sid": "AllObjectActions",
#         "Effect": "Allow",
#         "Action": "s3:*Object",
#         "Resource": [
#           aws_s3_bucket ,
#           "${aws_s3_bucket }/*"
#         ]
#       }
#     ]
#   })
# }

# resource "aws_iam_role_policy_attachment" "cortex-operator-config-s3-attach" {
#   role       = aws_iam_role.k8s-acc-node.name
#   policy_arn = aws_iam_policy.cortex-operator-config-s3.arn
# }
