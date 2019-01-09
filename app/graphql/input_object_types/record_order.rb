# frozen_string_literal: true

InputObjectTypes::RecordOrder = GraphQL::InputObjectType.define do
  name "RecordOrder"

  argument :field, !Types::Enum::RecordOrderField
  argument :direction, !Types::Enum::OrderDirection
end
