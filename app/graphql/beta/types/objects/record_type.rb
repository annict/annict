# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class RecordType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :user, Beta::Types::Objects::UserType, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :episode, Beta::Types::Objects::EpisodeType, null: false
        field :comment, String, null: true
        field :rating, Float, null: true
        field :rating_state, Beta::Types::Enums::RatingState, null: true
        field :modified, Boolean, null: false
        field :likes_count, Integer, null: false
        field :comments_count, Integer, null: false
        field :twitter_click_count, Integer, null: false
        field :facebook_click_count, Integer, null: false
        field :created_at, Beta::Types::Scalars::DateTime, null: false
        field :updated_at, Beta::Types::Scalars::DateTime, null: false

        def user
          Beta::RecordLoader.for(User).load(object.user_id)
        end

        def work
          Beta::RecordLoader.for(Work).load(object.work_id)
        end

        def episode
          Beta::RecordLoader.for(Episode).load(object.episode_id)
        end

        def comment
          object.body
        end

        def modified
          object.modify_body?
        end

        def created_at
          object.record.watched_at
        end

        private

        def record_promise
          Beta::RecordLoader.for(Record, column: :recordable_id, where: {recordable_type: "EpisodeRecord"}).load(object.id)
        end
      end
    end
  end
end
