# frozen_string_literal: true

class CastStruct < ApplicationStruct
  attribute :name, StructTypes::Strict::String
  attribute :name_en, StructTypes::Strict::String
  attribute :character, CharacterStruct
  attribute :person, PersonStruct
end
