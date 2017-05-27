# frozen_string_literal: true

InputObjectTypes::ProgramOrder = GraphQL::InputObjectType.define do
  name "ProgramOrder"

  argument :field, !EnumTypes::ProgramOrderField
  argument :direction, !EnumTypes::OrderDirection
end
