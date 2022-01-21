module "dynamodb_table" {
  source = "terraform-aws-modules/dynamodb-table/aws"

  billing_mode = "PAY_PER_REQUEST"

  attributes = [
    {
      name = "UserID"
      type = "S"
    },
    {
      name = "PostID"
      type = "S"
    }
  ]

  name      = "Posts"
  hash_key  = "UserID"
  range_key = "PostID"

  tags = {
    Environment = var.environment
  }
}
