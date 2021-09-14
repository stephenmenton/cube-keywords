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
