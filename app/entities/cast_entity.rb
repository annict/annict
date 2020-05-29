# frozen_string_literal: true

class CastEntity < ApplicationEntity
  local_attributes :accurate_name

  attribute? :accurate_name, Types::String
  attribute? :accurate_name_en, Types::String.optional
  attribute? :character, CharacterEntity
  attribute? :person, PersonEntity
end
