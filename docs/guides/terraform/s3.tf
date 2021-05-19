resource "aws_s3_bucket" "cortex_data_bucket" {
  bucket        = "${local.cortex_data_bucket_name}"
  acl           = "private"
  force_destroy = "true"
}

resource "aws_s3_bucket" "cortex_config_bucket" {
  bucket        = "${local.cortex_config_bucket_name}"
  acl           = "private"
  force_destroy = "true"
}
