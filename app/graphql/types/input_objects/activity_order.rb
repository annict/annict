# frozen_string_literal: true

module Types
  module InputObjects
    class ActivityOrder < Types::InputObjects::Base
      argument :field, Types::Enum::ActivityOrderField, required: true
      argument :direction, Types::Enum::OrderDirection, required: true
    end
  end
end
