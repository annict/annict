# frozen_string_literal: true

EnumTypes::OrderDirection = GraphQL::EnumType.define do
  name "OrderDirection"

  value "ASC", ""
  value "DESC", ""
end
