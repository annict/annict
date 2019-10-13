# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class EpisodeType < Canary::Types::Objects::Base
        description "エピソード情報"

        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :number, Integer, null: true
        field :number_text, String, null: true
        field :sort_number, Integer, null: false
        field :title, String, null: true
        field :satisfaction_rate, Float,
          null: true,
          description: "満足度"
        field :episode_records_count, Integer, null: false
        field :episode_record_bodies_count, Integer, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :prev_episode, Canary::Types::Objects::EpisodeType, null: true
        field :next_episode, Canary::Types::Objects::EpisodeType, null: true
        field :viewer_did_track, Boolean, null: false
        field :viewer_records_count, Integer, null: false

        field :episode_records, Canary::Types::Objects::EpisodeRecordType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::EpisodeRecordOrder, required: false
          argument :has_body, Boolean, required: false
        end

        def number
          object.raw_number
        end

        def number_text
          object.number
        end

        def episode_record_bodies_count
          object.episode_record_bodies_count
        end

        def work
          RecordLoader.for(Work).load(object.work_id)
        end

        def prev_episode
          RecordLoader.for(Episode).load(object.prev_episode_id)
        end

        def next_episode
          ForeignKeyLoader.for(Episode, :prev_episode_id).load([object.id]).then(&:first)
        end

        def viewer_did_track
          context[:viewer].tracked?(object)
        end

        def viewer_records_count
          context[:viewer].episode_records_count_in(object)
        end

        def episode_records(order_by: nil, has_body: nil)
          SearchEpisodeRecordsQuery.new(
            object.episode_records,
            order_by: order_by,
            has_body: has_body
          ).call
        end
      end
    end
  end
end
