# frozen_string_literal: true

Types::RatingStateEnum = GraphQL::EnumType.define do
  name "RatingState"

  value "GOOD", ""
  value "BAD", ""
end
