# frozen_string_literal: true

ScalarTypes::DateTime = GraphQL::ScalarType.define do
  name "DateTime"

  coerce_input ->(value, _ctx) { Time.zone.parse(value) }
  coerce_result ->(value, _ctx) { value.iso8601 }
end
