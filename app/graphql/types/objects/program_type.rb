# frozen_string_literal: true

module Types
  module Objects
    class ProgramType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :channel, Types::Objects::ChannelType, null: false
      field :episode, Types::Objects::EpisodeType, null: false
      field :work, Types::Objects::WorkType, null: false
      field :started_at, Types::Scalars::DateTime, null: false
      field :sc_pid, Integer, null: true
      field :state, Types::Enums::ProgramState, null: false
      field :rebroadcast, Boolean, null: false

      def channel
        RecordLoader.for(Channel).load(object.channel_id)
      end

      def episode
        RecordLoader.for(Episode).load(object.episode_id)
      end

      def work
        RecordLoader.for(Anime).load(object.anime_id)
      end

      def state
        (object.not_deleted? ? "published" : "hidden").upcase
      end
    end
  end
end
