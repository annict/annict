# frozen_string_literal: true

Connections::ActivityConnection =
  ObjectTypes::Activity.define_connection(edge_type: Edges::ActivityEdge) do
    name "ActivityConnection"
  end
