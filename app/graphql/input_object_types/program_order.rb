# frozen_string_literal: true

InputObjectTypes::ProgramOrder = GraphQL::InputObjectType.define do
  name "ProgramOrder"

  argument :field, !Types::Enum::ProgramOrderField
  argument :direction, !Types::Enum::OrderDirection
end
