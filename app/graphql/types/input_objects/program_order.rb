# frozen_string_literal: true

module Types
  module InputObjects
    class ProgramOrder < Types::InputObjects::Base
      argument :field, Types::Enums::ProgramOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
