# frozen_string_literal: true

InputObjectTypes::RecordOrder = GraphQL::InputObjectType.define do
  name "RecordOrder"

  argument :field, !EnumTypes::RecordOrderField
  argument :direction, !EnumTypes::OrderDirection
end
