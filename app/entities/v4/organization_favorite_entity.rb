# frozen_string_literal: true

module V4
  class OrganizationFavoriteEntity < V4::ApplicationEntity
    attribute? :organization, OrganizationEntity
    attribute? :watched_anime_count, Types::Integer

    def self.from_node(organization_favorite_node)
      attrs = {}

      if organization_node = organization_favorite_node["organization"]
        attrs[:organization] = V4::OrganizationEntity.from_node(organization_node)
      end

      if watched_anime_count = organization_favorite_node["watchedAnimeCount"]
        attrs[:watched_anime_count] = watched_anime_count
      end

      new attrs
    end
  end
end
