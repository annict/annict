# frozen_string_literal: true

EnumTypes::RatingState = GraphQL::EnumType.define do
  name "RatingState"

  value "GREAT", ""
  value "GOOD", ""
  value "AVERAGE", ""
  value "BAD", ""
end
