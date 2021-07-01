# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class EpisodeType < Canary::Types::Objects::Base
        description "エピソード情報"

        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: false
        field :raw_number, Float, null: true
        field :number, String, null: true
        field :number_en, String, null: true
        field :sort_number, Integer, null: false
        field :title, String, null: true
        field :title_en, String, null: true
        field :satisfaction_rate, Float,
          null: true,
          description: "満足度"
        field :viewer_tracked, Boolean, null: false
        field :viewer_tracked_in_current_status, Boolean, null: false
        field :episode_records_count, Integer, null: false
        field :commented_episode_records_count, Integer, null: false
        field :viewer_records_count, Integer, null: false
        field :anime, Canary::Types::Objects::AnimeType, null: false
        field :prev_episode, Canary::Types::Objects::EpisodeType, null: true
        field :next_episode, Canary::Types::Objects::EpisodeType, null: true

        field :records, Canary::Types::Objects::RecordType.connection_type, null: false, resolver: Canary::Resolvers::Records do
          argument :has_body, Boolean, required: false
          argument :by_following, Boolean, required: false
          argument :order_by, Canary::Types::InputObjects::RecordOrder, required: false
        end

        def anime
          RecordLoader.for(Anime).load(object.work_id)
        end

        def prev_episode
          RecordLoader.for(Episode).load(object.prev_episode_id)
        end

        def next_episode
          ForeignKeyLoader.for(Episode, :prev_episode_id).load([object.id]).then(&:first)
        end

        def viewer_tracked
          context[:viewer].tracked?(object)
        end

        def viewer_tracked_in_current_status
          RecordLoader.for(LibraryEntry, column: :work_id, where: {user_id: context[:viewer].id}).load(object.work_id).then do |le|
            le.watched_episode_ids.include?(object.id)
          end
        end

        def viewer_records_count
          context[:viewer].episode_records_count_in(object)
        end
      end
    end
  end
end
