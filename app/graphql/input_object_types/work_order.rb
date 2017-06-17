# frozen_string_literal: true

InputObjectTypes::WorkOrder = GraphQL::InputObjectType.define do
  name "WorkOrder"

  argument :field, !EnumTypes::WorkOrderField
  argument :direction, !EnumTypes::OrderDirection
end
