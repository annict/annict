# frozen_string_literal: true

module Types
  module InputObjects
    class ActivityOrder < Types::InputObjects::Base
      argument :field, Types::Enums::ActivityOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
