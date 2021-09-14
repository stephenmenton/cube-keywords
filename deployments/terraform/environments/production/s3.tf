resource "aws_s3_bucket" "cube_keywords" {
  bucket        = "cube-keywords"
  acl           = "private"
  force_destroy = false

  versioning {
    enabled = false
  }
}

resource "aws_s3_bucket_policy" "cube_keywords_lambda" {
  bucket = aws_s3_bucket.cube_keywords.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "LambdaRead",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "${aws_iam_role.cube_keywords_lambda.arn}"
        },
        "Action" : [
          "s3:GetObject",
          "s3:GetObjectVersion"
        ],
        "Resource" : "${aws_s3_bucket.cube_keywords.arn}/*"
      }
    ]
  })
}
