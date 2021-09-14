terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
  backend "s3" {
    bucket = "terraform-dad6179d51d50997fb57c92d"
    key    = "stephenmenton/cube_keywords"
    region = "us-west-2"
  }
}
