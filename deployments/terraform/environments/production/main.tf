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

data "archive_file" "zip" {
  type        = "zip"
  source_file = "../../../../main"
  output_path = "main.zip"
}

resource "aws_lambda_function" "cube_keywords" {
  function_name    = var.lambda_function_name
  filename         = "main.zip"
  handler          = "main"
  source_code_hash = "data.archive_file.zip.output_base64sha256"
  role             = aws_iam_role.cube_keywords_lambda.arn
  runtime          = "go1.x"
  memory_size      = 2048
  timeout          = 90
}

resource "aws_iam_role" "cube_keywords_lambda" {
  name = "${var.lambda_function_name}_lambda"
  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      }
    ]
  })
  managed_policy_arns = [aws_iam_policy.cube_keywords_lambda.arn]
  path                = "/service-role/"
}

// TODO: determine and remove unneeded s3 read permissions
resource "aws_iam_policy" "cube_keywords_lambda" {
  name        = "${var.lambda_function_name}_lambda"
  path        = "/"
  description = "primary /cube/{cube} endpoint"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "logsa",
        "Effect" : "Allow",
        "Action" : [
          "logs:CreateLogGroup"
        ],
        "Resource" : [
          "arn:aws:logs:us-west-2:947662671000:*"
        ]
      },
      {
        "Sid" : "logsb",
        "Effect" : "Allow",
        "Action" : [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        "Resource" : "arn:aws:logs:us-west-2:947662671000:log-group:/aws/lambda/${var.lambda_function_name}:*"
      },
      {
        "Sid" : "s3a",
        "Effect" : "Allow",
        "Action" : [
          "s3:GetObject"
        ],
        "Resource" : [
          "arn:aws:s3:::cube-keywords"
        ]
      },
      {
        "Sid" : "s3b",
        "Effect" : "Allow",
        "Action" : [
          "s3:GetAccessPoint",
          "s3:GetAccountPublicAccessBlock"
        ],
        "Resource" : "*"
      }
    ]
  })
}

resource "aws_apigatewayv2_api" "cube_keywords" {
  name          = var.lambda_function_name
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "cube_keywords_default" {
  api_id      = aws_apigatewayv2_api.cube_keywords.id
  name        = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "cube_keywords_lambda" {
  api_id           = aws_apigatewayv2_api.cube_keywords.id
  integration_type = "AWS_PROXY"
  integration_uri  = aws_lambda_function.cube_keywords.invoke_arn
}

resource "aws_apigatewayv2_route" "cube_text" {
  api_id    = aws_apigatewayv2_api.cube_keywords.id
  route_key = "ANY /cube/{cube}"
  target    = "integrations/${aws_apigatewayv2_integration.cube_keywords_lambda.id}"
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.cube_keywords.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.cube_keywords.execution_arn}/*/*"
}