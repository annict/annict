# frozen_string_literal: true

class EpisodeRecordEntity < ApplicationEntity
  attribute? :id, Types::Integer
  attribute? :rating_state, Types::RecordRatingStateKinds.optional
  attribute? :body, Types::String.optional
  attribute? :body_html, Types::String.optional
  attribute? :likes_count, Types::Integer
  attribute? :comments_count, Types::Integer
  attribute? :user, UserEntity
  attribute? :record, RecordEntity
  attribute? :work, WorkEntity
  attribute? :episode, EpisodeEntity
end
