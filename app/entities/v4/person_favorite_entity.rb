# frozen_string_literal: true

module Deprecated
  class PersonFavoriteEntity < Deprecated::ApplicationEntity
    attribute? :person, Deprecated::PersonEntity
    attribute? :watched_anime_count, Types::Integer

    def self.from_node(person_favorite_node)
      attrs = {}

      if person_node = person_favorite_node["person"]
        attrs[:person] = Deprecated::PersonEntity.from_node(person_node)
      end

      if watched_anime_count = person_favorite_node["watchedAnimeCount"]
        attrs[:watched_anime_count] = watched_anime_count
      end

      new attrs
    end
  end
end
