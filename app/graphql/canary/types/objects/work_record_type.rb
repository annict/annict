# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class WorkRecordType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :record, Canary::Types::Objects::RecordType, null: false
        field :title, String, null: true
        field :body, String, null: false
        field :body_html, String, null: false
        field :rating_overall_state, Canary::Types::Enums::RatingState, null: true
        field :rating_animation_state, Canary::Types::Enums::RatingState, null: true
        field :rating_music_state, Canary::Types::Enums::RatingState, null: true
        field :rating_story_state, Canary::Types::Enums::RatingState, null: true
        field :rating_character_state, Canary::Types::Enums::RatingState, null: true
        field :viewer_did_like, Boolean, null: false
        field :likes_count, Integer, null: false
        field :modified_at, Canary::Types::Scalars::DateTime, null: true
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :updated_at, Canary::Types::Scalars::DateTime, null: false

        def user
          Canary::RecordLoader.for(User).load(object.user_id)
        end

        def work
          Canary::RecordLoader.for(Work).load(object.work_id)
        end

        def record
          Canary::RecordLoader.for(Record).load(object.record_id)
        end

        def title
          object.title.presence || I18n.t("noun.record_of_work", work_title: object.work.local_title)
        end

        def body_html
          render_markdown(object.body)
        end

        def viewer_did_like
          return false unless context[:viewer]

          context[:viewer].like?(object)
        end
      end
    end
  end
end
