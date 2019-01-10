# frozen_string_literal: true

module Types
  module InputObjects
    class WorkOrder < Types::InputObjects::Base
      argument :field, Types::Enum::WorkOrderField, required: true
      argument :direction, !Types::Enum::OrderDirection, required: true
    end
  end
end
