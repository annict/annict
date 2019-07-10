# frozen_string_literal: true

class CastStruct < ApplicationStruct
  attribute :character, CharacterStruct
  attribute :person, PersonStruct
end
