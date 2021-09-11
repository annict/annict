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
          record_promise.then do |record|
            Beta::RecordLoader.for(User).load(record.user_id)
          end
        end

        def work
          record_promise.then do |record|
            Beta::RecordLoader.for(Work).load(record.work_id)
          end
        end

        def title
          ""
        end

        def body
          record_promise.then(&:body)
        end

        def rating_overall_state
          record_promise.then(&:rating)
        end

        def rating_animation_state
          record_promise.then(&:animation_rating)
        end

        def rating_music_state
          record_promise.then(&:music_rating)
        end

        def rating_story_state
          record_promise.then(&:story_rating)
        end

        def rating_character_state
          record_promise.then(&:character_rating)
        end

        def likes_count
          record_promise.then(&:likes_count)
        end

        def impressions_count
          record_promise.then(&:impressions_count)
        end

        def modified_at
          record_promise.then(&:modified_at)
        end

        def created_at
          record_promise.then(&:watched_at)
        end

        def updated_at
          record_promise.then(&:updated_at)
        end

        private

        def record_promise
          Beta::RecordLoader.for(Record, column: :recordable_id, where: {recordable_type: "WorkRecord"}).load(object.id)
        end
      end
    end
  end
end
