# frozen_string_literal: true

module Types
  module InputObjects
    class CharacterOrder < Types::InputObjects::Base
      argument :field, Types::Enums::CharacterOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
