# frozen_string_literal: true

InputObjectTypes::WorkOrder = GraphQL::InputObjectType.define do
  name "WorkOrder"

  argument :field, !Types::Enum::WorkOrderField
  argument :direction, !Types::Enum::OrderDirection
end
