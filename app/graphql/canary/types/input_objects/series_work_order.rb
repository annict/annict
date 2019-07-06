# frozen_string_literal: true

module Types
  module InputObjects
    class SeriesWorkOrder < Types::InputObjects::Base
      argument :field, Types::Enums::SeriesWorkOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
