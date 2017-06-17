# frozen_string_literal: true

ScalarTypes::DateTime = GraphQL::ScalarType.define do
  name "DateTime"

  coerce_input ->(value) { Time.zone.parse(value) }
  coerce_result ->(value) { value.iso8601 }
end
