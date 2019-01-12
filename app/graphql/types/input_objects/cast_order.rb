# frozen_string_literal: true

module Types
  module InputObjects
    class CastOrder < Types::InputObjects::Base
      argument :field, Types::Enums::CastOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
