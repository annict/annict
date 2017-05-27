# frozen_string_literal: true

EnumTypes::RatingState = GraphQL::EnumType.define do
  name "RatingState"

  value "GOOD", ""
  value "BAD", ""
end
