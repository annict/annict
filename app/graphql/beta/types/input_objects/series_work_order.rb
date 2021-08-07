# frozen_string_literal: true

module Beta
  module Types
    module InputObjects
      class SeriesWorkOrder < Beta::Types::InputObjects::Base
        argument :field, Beta::Types::Enums::SeriesWorkOrderField, required: true
        argument :direction, Beta::Types::Enums::OrderDirection, required: true
      end
    end
  end
end
