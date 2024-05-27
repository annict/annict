# typed: false
# frozen_string_literal: true

module Beta
  module Types
    module InputObjects
      class CharacterOrder < Beta::Types::InputObjects::Base
        argument :field, Beta::Types::Enums::CharacterOrderField, required: true
        argument :direction, Beta::Types::Enums::OrderDirection, required: true
      end
    end
  end
end
