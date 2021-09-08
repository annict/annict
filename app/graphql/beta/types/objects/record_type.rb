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
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            Beta::RecordLoader.for(User).load(records.first.user_id)
          end
        end

        def work
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            Beta::RecordLoader.for(Work).load(records.first.work_id)
          end
        end

        def episode
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            Beta::RecordLoader.for(Episode).load(records.first.episode_id)
          end
        end

        def comment
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            records.first.body
          end
        end

        def rating
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            records.first.advanced_rating
          end
        end

        def rating_state
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            records.first.rating
          end
        end

        def modified
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            !records.first.modified_at.nil?
          end
        end

        def likes_count
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            records.first.likes_count
          end
        end

        def comments_count
          Beta::ForeignKeyLoader.for(Record, :recordable_id).load([object.id]).then do |records|
            records.first.comments_count
          end
        end

        def twitter_click_count
          0
        end

        def facebook_click_count
          0
        end
      end
    end
  end
end
