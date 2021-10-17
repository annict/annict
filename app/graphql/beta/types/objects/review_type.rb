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
          Beta::RecordLoader.for(User).load(object.user_id)
        end

        def work
          Beta::RecordLoader.for(Work).load(object.work_id)
        end

        def title
          object.title.presence || I18n.t("noun.record_of_work", work_title: object.work.local_title)
        end

        def created_at
          object.record.watched_at
        end

        private

        def record_promise
          Beta::RecordLoader.for(Record, column: :recordable_id, where: {recordable_type: "WorkRecord"}).load(object.id)
        end
      end
    end
  end
end
