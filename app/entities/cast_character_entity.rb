# frozen_string_literal: true

class CastCharacterEntity < ApplicationEntity
  local_attributes :name

  attribute? :id, Types::Integer
  attribute? :name, Types::String
  attribute? :name_en, Types::String.optional
end
