# frozen_string_literal: true

class PersonFavoriteEntity < ApplicationEntity
  attribute? :person, PersonEntity
  attribute? :watched_works_count, Types::Integer

  def self.from_node(person_favorite_node)
    attrs = {}

    if person_node = person_favorite_node["person"]
      attrs[:person] = PersonEntity.from_node(person_node)
    end

    if watched_works_count = person_favorite_node["watchedWorksCount"]
      attrs[:watched_works_count] = watched_works_count
    end

    new attrs
  end
end
