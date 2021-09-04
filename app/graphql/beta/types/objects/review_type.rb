# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class ReviewType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :user, Beta::Types::Objects::UserType, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :title, String, null: true
        field :body, String, null: false
        field :rating_overall_state, Beta::Types::Enums::RatingState, null: true
        field :rating_animation_state, Beta::Types::Enums::RatingState, null: true
        field :rating_music_state, Beta::Types::Enums::RatingState, null: true
        field :rating_story_state, Beta::Types::Enums::RatingState, null: true
        field :rating_character_state, Beta::Types::Enums::RatingState, null: true
        field :likes_count, Integer, null: false
        field :impressions_count, Integer, null: false
        field :modified_at, Beta::Types::Scalars::DateTime, null: true
        field :created_at, Beta::Types::Scalars::DateTime, null: false
        field :updated_at, Beta::Types::Scalars::DateTime, null: false

        def user
          Beta::RecordLoader.for(User).load(object.record.user_id)
        end

        def work
          Beta::RecordLoader.for(Work).load(object.record.work_id)
        end

        def title
          ""
        end

        def body
          object.record.body
        end

        def rating_overall_state
          object.record.rating
        end

        def rating_animation_state
          object.record.animation_rating
        end

        def rating_music_state
          object.record.music_rating
        end

        def rating_story_state
          object.record.story_rating
        end

        def rating_character_state
          object.record.character_rating
        end

        def likes_count
          object.record.likes_count
        end

        def impressions_count
          object.record.impressions_count
        end

        def modified_at
          object.record.modified_at
        end

        def created_at
          object.record.watched_at
        end

        def updated_at
          object.record.updated_at
        end
      end
    end
  end
end
