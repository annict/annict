# frozen_string_literal: true

module Types
  module Objects
    class ReviewType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :user, Types::Objects::UserType, null: false
      field :work, Types::Objects::WorkType, null: false
      field :title, String, null: true
      field :body, String, null: false
      field :rating_overall_state, Types::Enums::RatingState, null: true
      field :rating_animation_state, Types::Enums::RatingState, null: true
      field :rating_music_state, Types::Enums::RatingState, null: true
      field :rating_story_state, Types::Enums::RatingState, null: true
      field :rating_character_state, Types::Enums::RatingState, null: true
      field :likes_count, Integer, null: false
      field :impressions_count, Integer, null: false
      field :modified_at, Types::Scalars::DateTime, null: true
      field :created_at, Types::Scalars::DateTime, null: false
      field :updated_at, Types::Scalars::DateTime, null: false

      def user
        RecordLoader.for(User).load(object.user_id)
      end

      def work
        RecordLoader.for(Anime).load(object.work_id)
      end

      def title
        object.title.presence || I18n.t("noun.record_of_work", work_title: object.work.local_title)
      end
    end
  end
end
