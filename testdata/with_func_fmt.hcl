target {
block = base64encode(<<-EOT
                                  {
"body":
                                    {
"key1": "aaa",
                                     "key2": "bbb",
                                     "key3": "ccc"
                                   }
                                  }
                                EOT
)
}
