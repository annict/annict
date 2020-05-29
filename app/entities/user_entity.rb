# frozen_string_literal: true

class UserEntity < ApplicationEntity
  attribute? :id, Types::Integer
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
  attribute? :plan_to_watch_works_count, Types::Integer
  attribute? :watching_works_count, Types::Integer
  attribute? :completed_works_count, Types::Integer
  attribute? :on_hold_works_count, Types::Integer
  attribute? :dropped_works_count, Types::Integer
  attribute? :character_favorites_count, Types::Integer
  attribute? :person_favorites_count, Types::Integer
  attribute? :organization_favorites_count, Types::Integer
  attribute? :created_at, Types::Params::Time
  attribute? :character_favorites, Types::Array.of(CharacterFavoriteEntity)
  attribute? :cast_favorites, Types::Array.of(PersonFavoriteEntity)
  attribute? :staff_favorites, Types::Array.of(PersonFavoriteEntity)
  attribute? :organization_favorites, Types::Array.of(OrganizationFavoriteEntity)

  def days_from_started(time_zone)
    ((Time.zone.now.in_time_zone(time_zone) - created_at.in_time_zone(time_zone)) / 86_400).ceil
  end
end
