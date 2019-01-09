# frozen_string_literal: true

InputObjectTypes::ReviewOrder = GraphQL::InputObjectType.define do
  name "ReviewOrder"

  argument :field, !Types::Enum::ReviewOrderField
  argument :direction, !Types::Enum::OrderDirection
end
