# frozen_string_literal: true

EnumTypes::ReviewOrderField = GraphQL::EnumType.define do
  name "ReviewOrderField"

  value "CREATED_AT", ""
  value "LIKES_COUNT", ""
end
