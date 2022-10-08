# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class EpisodeType < Beta::Types::Objects::Base
        description "An episode of a work"

        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :number, Integer, null: true
        field :number_text, String, null: true
        field :sort_number, Integer, null: false
        field :title, String, null: true
        field :satisfaction_rate, Float, null: true
        field :records_count, Integer, null: false
        field :record_comments_count, Integer, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :prev_episode, Beta::Types::Objects::EpisodeType, null: true
        field :next_episode, Beta::Types::Objects::EpisodeType, null: true
        field :viewer_did_track, Boolean, null: false
        field :viewer_records_count, Integer, null: false

        field :records, Beta::Types::Objects::RecordType.connection_type, null: true do
          argument :order_by, Beta::Types::InputObjects::RecordOrder, required: false
          argument :has_comment, Boolean, required: false
        end

        def number
          object.raw_number
        end

        def number_text
          object.number
        end

        def records_count
          object.episode_records_count
        end

        def record_comments_count
          object.episode_record_bodies_count
        end

        def work
          Beta::RecordLoader.for(Work).load(object.work_id)
        end

        def prev_episode
          Beta::RecordLoader.for(Episode).load(object.prev_episode_id)
        end

        def next_episode
          Beta::ForeignKeyLoader.for(Episode, :prev_episode_id).load([object.id]).then(&:first)
        end

        def viewer_did_track
          context[:viewer].tracked?(object)
        end

        def viewer_records_count
          context[:viewer].episode_records_count_in(object)
        end

        def records(order_by: nil, has_comment: nil)
          Deprecated::SearchEpisodeRecordsQuery.new(
            object.episode_records,
            order_by: order_by,
            has_body: has_comment
          ).call
        end
      end
    end
  end
end
