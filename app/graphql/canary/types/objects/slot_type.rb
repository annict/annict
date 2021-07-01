# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class SlotType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: false
        field :channel, Canary::Types::Objects::ChannelType, null: false
        field :episode, Canary::Types::Objects::EpisodeType, null: true
        field :anime, Canary::Types::Objects::AnimeType, null: false
        field :started_at, Canary::Types::Scalars::DateTime, null: false
        field :sc_pid, Integer, null: true
        field :rebroadcast, Boolean, null: false

        def channel
          RecordLoader.for(Channel).load(object.channel_id)
        end

        def episode
          RecordLoader.for(Episode).load(object.episode_id)
        end

        def anime
          RecordLoader.for(Anime).load(object.work_id)
        end
      end
    end
  end
end
