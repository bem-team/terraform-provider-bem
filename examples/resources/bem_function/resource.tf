resource "bem_function" "example_function" {
  path_function_name = "functionName"
  type = "transform"
  display_name = "displayName"
  output_schema = {

  }
  output_schema_name = "outputSchemaName"
  tabular_chunking_enabled = true
  tags = ["string"]
}
