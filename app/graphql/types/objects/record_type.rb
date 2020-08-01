# frozen_string_literal: true

module Types
  module Objects
    class RecordType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :user, Types::Objects::UserType, null: false
      field :work, Types::Objects::WorkType, null: false
      field :episode, Types::Objects::EpisodeType, null: false
      field :comment, String, null: true
      field :rating, Float, null: true
      field :rating_state, Types::Enums::RatingState, null: true
      field :modified, Boolean, null: false
      field :likes_count, Integer, null: false
      field :comments_count, Integer, null: false
      field :twitter_click_count, Integer, null: false
      field :facebook_click_count, Integer, null: false
      field :created_at, Types::Scalars::DateTime, null: false
      field :updated_at, Types::Scalars::DateTime, null: false

      def user
        RecordLoader.for(User).load(object.user_id)
      end

      def work
        RecordLoader.for(Work).load(object.anime_id)
      end

      def episode
        RecordLoader.for(Episode).load(object.episode_id)
      end

      def comment
        object.body
      end

      def modified
        object.modify_body?
      end
    end
  end
end
