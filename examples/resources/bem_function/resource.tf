resource "bem_function" "example_function" {
  type = "extract"
  display_name = "displayName"
  enable_bounding_boxes = true
  output_schema = {

  }
  output_schema_name = "outputSchemaName"
  pre_count = true
  tabular_chunking_enabled = true
  tags = ["string"]
}
