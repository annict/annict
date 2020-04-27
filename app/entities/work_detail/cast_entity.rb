# frozen_string_literal: true

module WorkDetail
  class CastEntity < ApplicationEntity
    local_attributes :accurate_name

    attribute :accurate_name, Types::String
    attribute :accurate_name_en, Types::String.optional
    attribute :character, CastCharacterEntity
    attribute :person do
      attribute :id, Types::Integer
    end
  end
end
