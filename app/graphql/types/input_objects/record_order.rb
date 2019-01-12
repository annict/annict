# frozen_string_literal: true

module Types
  module InputObjects
    class RecordOrder < Types::InputObjects::Base
      argument :field, Types::Enums::RecordOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
