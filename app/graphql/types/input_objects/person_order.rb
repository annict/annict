# frozen_string_literal: true

module Types
  module InputObjects
    class PersonOrder < Types::InputObjects::Base
      argument :field, Types::Enums::PersonOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
