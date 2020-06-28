# frozen_string_literal: true

class CharacterEntity < ApplicationEntity
  local_attributes :name

  attribute? :database_id, Types::Integer
  attribute? :name, Types::String
  attribute? :name_en, Types::String.optional
  attribute? :series, SeriesEntity.optional
end
