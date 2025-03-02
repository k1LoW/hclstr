resource "aws_iam_policy" "allow_dynamodb_table_post" {
  name   = "allow_post"
  policy = <<-EOT
{
"Version": "2012-10-17",
  "Statement": {
"Effect": "Allow",
         "Action": "dynamodb:*",
"Resource": "${aws_dynamodb_table.post.arn}"
}
}
EOT
}
