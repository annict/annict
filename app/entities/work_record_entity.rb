# frozen_string_literal: true

class WorkRecordEntity < ApplicationEntity
  attribute :type, Types::Value("work_record")

  attribute? :id, Types::Integer
  attribute? :rating_animation_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_music_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_story_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_character_state, Types::RecordRatingStateKinds.optional
  attribute? :rating_overall_state, Types::RecordRatingStateKinds.optional
  attribute? :body, Types::String.optional
  attribute? :body_html, Types::String.optional
  attribute? :likes_count, Types::Integer
  attribute? :created_at, Types::Time
  attribute? :modified_at, Types::Time.optional
  attribute? :user, UserEntity
  attribute? :record, RecordEntity
  attribute? :work, WorkEntity
end
