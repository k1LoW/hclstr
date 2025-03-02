target {
block = <<-EOT
%{ for ip in aws_instance.example[*].private_ip ~}
server ${ip}
%{ endfor ~}
EOT
}
