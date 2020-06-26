# frozen_string_literal: true

class WorkRecordEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :database_id, Types::Integer
  attribute? :rating_animation_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_music_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_story_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_character_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_overall_state, Types::RecordRatingStateKinds.optional
  attribute? :body, Types::String.optional
  attribute? :body_html, Types::String.optional
  attribute? :likes_count, Types::Integer
  attribute? :created_at, Types::Params::Time
  attribute? :user, UserEntity
  attribute? :record, RecordEntity
  attribute? :work, WorkEntity

  def self.from_node(work_record_node, user_node: nil)
    attrs = {}

    if database_id = work_record_node["annictId"]
      attrs[:database_id] = database_id
    end

    attrs[:rating_animation_state] = work_record_node["ratingAnimationState"]&.downcase
    attrs[:rating_music_state] = work_record_node["ratingMusicState"]&.downcase
    attrs[:rating_story_state] = work_record_node["ratingStoryState"]&.downcase
    attrs[:rating_character_state] = work_record_node["ratingCharacterState"]&.downcase
    attrs[:rating_overall_state] = work_record_node["ratingOverallState"]&.downcase

    if body = work_record_node["body"]
      attrs[:body] = body
    end

    if likes_count = work_record_node["likesCount"]
      attrs[:likes_count] = likes_count
    end

    if work_node = work_record_node["work"]
      attrs[:work] = WorkEntity.from_node(work_node)
    end

    if user_node = work_record_node["user"] || user_node
      attrs[:user] = UserEntity.from_node(user_node)
    end

    new attrs
  end
end
