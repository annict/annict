# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class LibraryEntryType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :user, Canary::Types::Objects::UserType, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :status, Canary::Types::Objects::StatusType, null: false
        field :next_episode, Canary::Types::Objects::EpisodeType, null: false
        field :created_at, GraphQL::Types::ISO8601DateTime, null: false

        field :tapped_episodes, Canary::Connections::EpisodeConnection, null: false, connection: true do
          argument :order_by, Canary::Types::InputObjects::EpisodeOrder, required: false
        end

        field :untapped_episodes, Canary::Connections::EpisodeConnection, null: false, connection: true do
          argument :order_by, Canary::Types::InputObjects::EpisodeOrder, required: false
        end

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def work
          RecordLoader.for(Work).load(object.work_id)
        end

        def status
          RecordLoader.for(Status).load(object.status_id)
        end

        def next_episode
          RecordLoader.for(Episode).load(object.next_episode_id)
        end

        def tapped_episodes(order_by: nil)
          order = build_order(order_by)

          BatchLoader::GraphQL.for(id: object.work_id, tapped_episode_ids: tapped_episode_ids).batch(default_value: []) do |hashes, loader|
            all_tapped_episode_ids = hashes.pluck(:tapped_episode_ids).flatten
            work_ids = hashes.pluck(:id)
            Episode.
              without_deleted.
              where(id: all_tapped_episode_ids).
              where(work_id: work_ids).
              order(order.field => order.direction).
              each do |episode|
              tapped_episode_ids = hashes.find { |hash| hash[:id] == episode.work_id }[:tapped_episode_ids]
              loader.call(id: episode.work_id, tapped_episode_ids: tapped_episode_ids) { |memo| memo << episode }
            end
          end
        end

        def untapped_episodes(order_by: nil)
          order = build_order(order_by)

          BatchLoader::GraphQL.for(id: object.work_id, tapped_episode_ids: tapped_episode_ids).batch(default_value: []) do |hashes, loader|
            all_tapped_episode_ids = hashes.pluck(:tapped_episode_ids).flatten
            work_ids = hashes.pluck(:id)
            Episode.
              without_deleted.
              where.not(id: all_tapped_episode_ids).
              where(work_id: work_ids).
              order(order.field => order.direction).
              each do |episode|
                tapped_episode_ids = hashes.find { |hash| hash[:id] == episode.work_id }[:tapped_episode_ids]
                loader.call(id: episode.work_id, tapped_episode_ids: tapped_episode_ids) { |memo| memo << episode }
              end
          end
        end

        private

        def tapped_episode_ids
          object.watched_episode_ids
        end
      end
    end
  end
end
