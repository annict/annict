# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class EpisodeRecordType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :episode, Canary::Types::Objects::EpisodeType, null: false
        field :body, String, null: true
        field :body_html, String, null: true
        field :rating, Float, null: true
        field :rating_state, Canary::Types::Enums::RatingState, null: true
        field :modified, Boolean, null: false
        field :likes_count, Integer, null: false
        field :comments_count, Integer, null: false
        field :twitter_click_count, Integer, null: false
        field :facebook_click_count, Integer, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :updated_at, Canary::Types::Scalars::DateTime, null: false

        def body_html
          render_markdown(object.body)
        end

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def work
          RecordLoader.for(Work).load(object.work_id)
        end

        def episode
          RecordLoader.for(Episode).load(object.episode_id)
        end

        def modified
          object.modify_body?
        end
      end
    end
  end
end
