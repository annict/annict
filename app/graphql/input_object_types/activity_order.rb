# frozen_string_literal: true

InputObjectTypes::ActivityOrder = GraphQL::InputObjectType.define do
  name "ActivityOrder"

  argument :field, !EnumTypes::ActivityOrderField
  argument :direction, !EnumTypes::OrderDirection
end
