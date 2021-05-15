# frozen_string_literal: true

module Deprecated
  class CharacterFavoriteEntity < Deprecated::ApplicationEntity
    attribute? :character, CharacterEntity

    def self.from_node(character_favorite_node)
      attrs = {}

      if character_node = character_favorite_node["character"]
        attrs[:character] = Deprecated::CharacterEntity.from_node(character_node)
      end

      new attrs
    end
  end
end
