# frozen_string_literal: true

EnumTypes::RecordOrderField = GraphQL::EnumType.define do
  name "RecordOrderField"

  value "CREATED_AT", ""
  value "LIKES_COUNT", ""
end
