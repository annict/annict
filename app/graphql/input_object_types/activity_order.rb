# frozen_string_literal: true

InputObjectTypes::ActivityOrder = GraphQL::InputObjectType.define do
  name "ActivityOrder"

  argument :field, !Types::Enum::ActivityOrderField
  argument :direction, !Types::Enum::OrderDirection
end
