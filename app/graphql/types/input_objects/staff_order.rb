# frozen_string_literal: true

module Types
  module InputObjects
    class StaffOrder < Types::InputObjects::Base
      argument :field, Types::Enums::StaffOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
