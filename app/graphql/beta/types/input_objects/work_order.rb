# typed: false
# frozen_string_literal: true

module Beta
  module Types
    module InputObjects
      class WorkOrder < Beta::Types::InputObjects::Base
        argument :field, Beta::Types::Enums::WorkOrderField, required: true
        argument :direction, Beta::Types::Enums::OrderDirection, required: true
      end
    end
  end
end
