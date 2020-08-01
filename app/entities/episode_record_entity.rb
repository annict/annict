# frozen_string_literal: true

class EpisodeRecordEntity < ApplicationEntity
  attribute? :database_id, Types::Integer
  attribute? :rating_state, Types::RecordRatingStateKinds.optional
  attribute? :body, Types::String.optional
  attribute? :body_html, Types::String.optional
  attribute? :likes_count, Types::Integer
  attribute? :comments_count, Types::Integer
  attribute? :user, UserEntity
  attribute? :record, RecordEntity
  attribute? :work, AnimeEntity
  attribute? :episode, EpisodeEntity

  def self.from_node(episode_record_node, user_node: nil)
    attrs = {}

    if database_id = episode_record_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if rating_state = episode_record_node["ratingState"]
      attrs[:rating_state] = rating_state.downcase
    end

    if body = episode_record_node["body"]
      attrs[:body] = body
    end

    if likes_count = episode_record_node["likesCount"]
      attrs[:likes_count] = likes_count
    end

    if comments_count = episode_record_node["commentsCount"]
      attrs[:comments_count] = comments_count
    end

    if work_node = episode_record_node["work"]
      attrs[:work] = AnimeEntity.from_node(work_node)
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
