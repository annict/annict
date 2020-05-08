# frozen_string_literal: true

class PersonFavoriteEntity < ApplicationEntity
  attribute? :person, PersonEntity
  attribute? :watched_works_count, Types::Integer
end
