output "base_url" {
  description = "/cube/{cube} base URL"

  value = aws_apigatewayv2_stage.cube_keywords_default.invoke_url
}