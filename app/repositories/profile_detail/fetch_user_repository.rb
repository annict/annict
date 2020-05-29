# frozen_string_literal: true

module ProfileDetail
  class FetchUserRepository < ApplicationRepository
    def fetch(username:)
      result = execute(variables: { username: username })
      node = result.to_h.dig("data", "user")

      UserEntity.new(
        id: node["annictId"],
        username: node["username"],
        name: node["name"],
        description: node["description"],
        avatar_url: node["avatarUrl"],
        background_image_url: node["backgroundImageUrl"],
        display_supporter_badge: node["displaySupporterBadge"],
        records_count: node["recordsCount"],
        watching_works_count: node["watchingWorksCount"],
        following_count: node["followingCount"],
        followers_count: node["followersCount"],
        character_favorites_count: node["characterFavoritesCount"],
        person_favorites_count: node["personFavoritesCount"],
        organization_favorites_count: node["organizationFavoritesCount"],
        organization_favorites_count: node["organizationFavoritesCount"],
        created_at: node["createdAt"],
        character_favorites: character_favorites(node.dig("characterFavorites", "nodes")),
        cast_favorites: person_favorites(node.dig("castFavorites", "nodes")),
        staff_favorites: person_favorites(node.dig("staffFavorites", "nodes")),
        organization_favorites: organization_favorites(node.dig("organizationFavorites", "nodes"))
      )
    end

    private

    def character_favorites(nodes)
      nodes.map do |node|
        character = node["character"]
        series = character["series"]

        CharacterFavoriteEntity.new(
          character: CharacterEntity.new(
            id: character["annictId"],
            name: character["name"],
            name_en: character["nameEn"],
            series: build_series(series)
          )
        )
      end
    end

    def build_series(series)
      return unless series

      SeriesEntity.new(
        name: series["name"],
        name_en: series["nameEn"]
      )
    end

    def person_favorites(nodes)
      nodes.map do |node|
        person = node["person"]

        PersonFavoriteEntity.new(
          person: PersonEntity.new(
            id: person["annictId"],
            name: person["name"],
            name_en: person["nameEn"]
          ),
          watched_works_count: node["watchedWorksCount"]
        )
      end
    end

    def organization_favorites(nodes)
      nodes.map do |node|
        organization = node["organization"]

        OrganizationFavoriteEntity.new(
          organization: OrganizationEntity.new(
            id: organization["annictId"],
            name: organization["name"],
            name_en: organization["nameEn"]
          ),
          watched_works_count: node["watchedWorksCount"]
        )
      end
    end
  end
end
