# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class MultipleEpisodeRecordType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :work, Canary::Types::Objects::AnimeType, null: false
        field :records, Canary::Types::Objects::EpisodeRecordType.connection_type, null: true
        field :created_at, Canary::Types::Scalars::DateTime, null: false

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def work
          RecordLoader.for(Anime).load(object.anime_id)
        end

        def records
          ForeignKeyLoader.for(EpisodeRecord, :multiple_episode_record_id).load([object.id])
        end
      end
    end
  end
end
