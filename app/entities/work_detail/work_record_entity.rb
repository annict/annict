# frozen_string_literal: true

module WorkDetail
  class WorkRecordEntity < ApplicationEntity
    attribute :id, Types::Integer
    attribute :rating_animation_state, Types::RecordRatingStateKinds.optional
    attribute :rating_music_state, Types::RecordRatingStateKinds.optional
    attribute :rating_story_state, Types::RecordRatingStateKinds.optional
    attribute :rating_character_state, Types::RecordRatingStateKinds.optional
    attribute :rating_overall_state, Types::RecordRatingStateKinds.optional
    attribute :body_html, Types::String.optional
    attribute :likes_count, Types::Integer
    attribute :created_at, Types::DateTime
    attribute :modified_at, Types::DateTime.optional
    attribute :viewer_did_like, Types::Bool
    attribute :user do
      attribute :username, Types::String
      attribute :name, Types::String.optional
      attribute :avatar_url, Types::String
      attribute :is_supporter, Types::Bool
    end
    attribute :record do
      attribute :id, Types::Integer
    end
  end
end
