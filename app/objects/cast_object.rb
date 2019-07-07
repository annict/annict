# frozen_string_literal: true

class CastObject < ApplicationObject
  attribute :character, CharacterObject
  attribute :person, PersonObject
end
