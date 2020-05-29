# frozen_string_literal: true

class OrganizationFavoriteEntity < ApplicationEntity
  attribute? :organization, OrganizationEntity
  attribute? :watched_works_count, Types::Integer
end
