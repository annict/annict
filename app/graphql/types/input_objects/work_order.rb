# frozen_string_literal: true

module Types
  module InputObjects
    class WorkOrder < Types::InputObjects::Base
      argument :field, Types::Enums::WorkOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
