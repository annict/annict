# frozen_string_literal: true

InputObjectTypes::ReviewOrder = GraphQL::InputObjectType.define do
  name "ReviewOrder"

  argument :field, !EnumTypes::ReviewOrderField
  argument :direction, !EnumTypes::OrderDirection
end
