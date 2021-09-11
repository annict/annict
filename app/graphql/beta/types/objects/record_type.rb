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
          record_promise.then do |record|
            Beta::RecordLoader.for(User).load(record.user_id)
          end
        end

        def work
          record_promise.then do |record|
            Beta::RecordLoader.for(Work).load(record.work_id)
          end
        end

        def episode
          record_promise.then do |record|
            Beta::RecordLoader.for(Episode).load(record.episode_id)
          end
        end

        def comment
          record_promise.then(&:body)
        end

        def rating
          record_promise.then(&:advanced_rating)
        end

        def rating_state
          record_promise.then(&:rating)
        end

        def modified
          record_promise.then do |record|
            !record.modified_at.nil?
          end
        end

        def likes_count
          record_promise.then(&:likes_count)
        end

        def comments_count
          record_promise.then(&:comments_count)
        end

        def twitter_click_count
          0
        end

        def facebook_click_count
          0
        end

        private

        def record_promise
          Beta::RecordLoader.for(Record, column: :recordable_id, where: {recordable_type: "EpisodeRecord"}).load(object.id)
        end
      end
    end
  end
end
