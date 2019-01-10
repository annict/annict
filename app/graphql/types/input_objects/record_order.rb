# frozen_string_literal: true

module Types
  module InputObjects
    class RecordOrder < Types::InputObjects::Base
      argument :field, Types::Enum::RecordOrderField, required: true
      argument :direction, !Types::Enum::OrderDirection, required: true
    end
  end
end
