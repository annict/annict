# frozen_string_literal: true

class EpisodeRecordEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :rating, Types::RecordRatingState.optional
  attribute? :advanced_rating, Types::Float.optional
  attribute? :comments_count, Types::Integer
  attribute? :user, UserEntity
  attribute? :record, RecordEntity
  attribute? :anime, AnimeEntity
  attribute? :episode, EpisodeEntity

  def self.from_node(episode_record_node, user_node: nil)
    attrs = {}

    if database_id = episode_record_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if rating = episode_record_node["rating"]
      attrs[:rating] = rating.downcase
    end

    if advanced_rating = episode_record_node["advancedRating"]
      attrs[:advanced_rating] = advanced_rating
    end

    if comments_count = episode_record_node["commentsCount"]
      attrs[:comments_count] = comments_count
    end

    if anime_node = episode_record_node["anime"]
      attrs[:anime] = AnimeEntity.from_node(anime_node)
    end

    if episode_node = episode_record_node["episode"]
      attrs[:episode] = EpisodeEntity.from_node(episode_node)
    end

    if record_node = episode_record_node["record"]
      attrs[:record] = RecordEntity.from_node(record_node)
    end

    if user_node = episode_record_node["user"] || user_node
      attrs[:user] = UserEntity.from_node(user_node)
    end

    new attrs
  end
end
