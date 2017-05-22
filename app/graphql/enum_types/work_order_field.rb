# frozen_string_literal: true

EnumTypes::WorkOrderField = GraphQL::EnumType.define do
  name "WorkOrderField"

  value "CREATED_AT", ""
  value "SEASON", ""
  value "WATCHERS_COUNT", ""
end
