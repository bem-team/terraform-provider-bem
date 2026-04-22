resource "bem_workflow" "example_workflow" {
  main_node_name = "mainNodeName"
  name = "name"
  nodes = [{
    function = {
      id = "id"
      name = "name"
      version_num = 0
    }
    metadata = {

    }
    name = "name"
  }]
  connectors = [{
    name = "name"
    type = "paragon"
    connector_id = "connectorID"
    paragon = {
      configuration = {

      }
      integration = "integration"
    }
  }]
  display_name = "displayName"
  edges = [{
    destination_node_name = "destinationNodeName"
    source_node_name = "sourceNodeName"
    destination_name = "destinationName"
    metadata = {

    }
  }]
  tags = ["string"]
}
