# frozen_string_literal: true

class CharacterFavoriteEntity < ApplicationEntity
  attribute? :character, CharacterEntity

  def self.from_node(character_favorite_node)
    attrs = {}

    if character_node = character_favorite_node["character"]
      attrs[:character] = CharacterEntity.from_node(character_node)
    end

    new attrs
  end
end
