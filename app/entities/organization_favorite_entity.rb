# frozen_string_literal: true

class OrganizationFavoriteEntity < ApplicationEntity
  attribute? :organization, OrganizationEntity
  attribute? :watched_works_count, Types::Integer

  def self.from_node(organization_favorite_node)
    attrs = {}

    if organization_node = organization_favorite_node["organization"]
      attrs[:organization] = OrganizationEntity.from_node(organization_node)
    end

    if watched_works_count = organization_favorite_node["watchedWorksCount"]
      attrs[:watched_works_count] = watched_works_count
    end

    new attrs
  end
end
