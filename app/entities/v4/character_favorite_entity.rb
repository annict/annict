# frozen_string_literal: true

module V4
  class CharacterFavoriteEntity < V4::ApplicationEntity
    attribute? :character, CharacterEntity

    def self.from_node(character_favorite_node)
      attrs = {}

      if (character_node = character_favorite_node["character"])
        attrs[:character] = V4::CharacterEntity.from_node(character_node)
      end

      new attrs
    end
  end
end
