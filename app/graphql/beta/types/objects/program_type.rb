# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class ProgramType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :channel, Beta::Types::Objects::ChannelType, null: false
        field :episode, Beta::Types::Objects::EpisodeType, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :started_at, Beta::Types::Scalars::DateTime, null: false
        field :sc_pid, Integer, null: true
        field :state, Beta::Types::Enums::ProgramState, null: false
        field :rebroadcast, Boolean, null: false

        def channel
          Beta::RecordLoader.for(Channel).load(object.channel_id)
        end

        def episode
          Beta::RecordLoader.for(Episode).load(object.episode_id)
        end

        def work
          Beta::RecordLoader.for(Work).load(object.work_id)
        end

        def state
          (object.not_deleted? ? "published" : "hidden").upcase
        end
      end
    end
  end
end
