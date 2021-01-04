# frozen_string_literal: true

class UserEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :username, Types::String
  attribute? :name, Types::String.optional
  attribute? :description, Types::String.optional
  attribute? :url, Types::String.optional
  attribute? :avatar_url, Types::String.optional
  attribute? :background_image_url, Types::String.optional
  attribute? :display_supporter_badge, Types::Bool
  attribute? :records_count, Types::Integer
  attribute? :following_count, Types::Integer
  attribute? :followers_count, Types::Integer
  attribute? :plan_to_watch_anime_count, Types::Integer
  attribute? :watching_anime_count, Types::Integer
  attribute? :completed_anime_count, Types::Integer
  attribute? :on_hold_anime_count, Types::Integer
  attribute? :dropped_anime_count, Types::Integer
  attribute? :character_favorites_count, Types::Integer
  attribute? :person_favorites_count, Types::Integer
  attribute? :organization_favorites_count, Types::Integer
  attribute? :created_at, Types::Params::Time
  attribute? :character_favorites, Types::Array.of(CharacterFavoriteEntity)
  attribute? :cast_favorites, Types::Array.of(PersonFavoriteEntity)
  attribute? :staff_favorites, Types::Array.of(PersonFavoriteEntity)
  attribute? :organization_favorites, Types::Array.of(OrganizationFavoriteEntity)

  def self.from_node(user_node)
    attrs = {}

    if database_id = user_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if username = user_node["username"]
      attrs[:username] = username
    end

    if name = user_node["name"]
      attrs[:name] = name
    end

    if description = user_node["description"]
      attrs[:description] = description
    end

    if url = user_node["url"]
      attrs[:url] = url
    end

    if avatar_url = user_node["avatarUrl"]
      attrs[:avatar_url] = avatar_url
    end

    if background_image_url = user_node["backgroundImageUrl"]
      attrs[:background_image_url] = background_image_url
    end

    if display_supporter_badge = user_node["displaySupporterBadge"]
      attrs[:display_supporter_badge] = display_supporter_badge
    end

    if records_count = user_node["recordsCount"]
      attrs[:records_count] = records_count
    end

    if watching_anime_count = user_node["watchingAnimeCount"]
      attrs[:watching_anime_count] = watching_anime_count
    end

    if following_count = user_node["followingCount"]
      attrs[:following_count] = following_count
    end

    if followers_count = user_node["followersCount"]
      attrs[:followers_count] = followers_count
    end

    if character_favorites_count = user_node["characterFavoritesCount"]
      attrs[:character_favorites_count] = character_favorites_count
    end

    if person_favorites_count = user_node["personFavoritesCount"]
      attrs[:person_favorites_count] = person_favorites_count
    end

    if organization_favorites_count = user_node["organizationFavoritesCount"]
      attrs[:organization_favorites_count] = organization_favorites_count
    end

    if created_at = user_node["createdAt"]
      attrs[:created_at] = created_at
    end

    character_favorite_nodes = user_node.dig("characterFavorites", "nodes")
    attrs[:character_favorites] = (character_favorite_nodes || []).map do |character_favorite_node|
      CharacterFavoriteEntity.from_node(character_favorite_node)
    end

    cast_favorite_nodes = user_node.dig("castFavorites", "nodes")
    attrs[:cast_favorites] = (cast_favorite_nodes || []).map do |cast_favorite_node|
      PersonFavoriteEntity.from_node(cast_favorite_node)
    end

    staff_favorite_nodes = user_node.dig("staffFavorites", "nodes")
    attrs[:staff_favorites] = (staff_favorite_nodes || []).map do |staff_favorite_node|
      PersonFavoriteEntity.from_node(staff_favorite_node)
    end

    organization_favorite_nodes = user_node.dig("organizationFavorites", "nodes")
    attrs[:organization_favorites] = (organization_favorite_nodes || []).map do |organization_favorite_node|
      OrganizationFavoriteEntity.from_node(organization_favorite_node)
    end

    new attrs
  end

  def self.from_model(user)
    extend Imgix::Rails::UrlHelper
    extend ImageHelper

    new(
      database_id: user.id,
      username: user.username,
      name: user.profile.name,
      avatar_url: api_user_avatar_url(user.profile, "size200"),
      background_image_url: ann_api_assets_background_image_url(user.profile),
      display_supporter_badge: user.supporter? && !user.setting.hide_supporter_badge?
    )
  end

  def days_from_started(time_zone)
    ((Time.zone.now.in_time_zone(time_zone) - created_at.in_time_zone(time_zone)) / 86_400).ceil
  end
end
