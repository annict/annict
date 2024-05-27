# typed: false
# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class MultipleRecordType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :user, Beta::Types::Objects::UserType, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :records, Beta::Types::Objects::RecordType.connection_type, null: true
        field :created_at, Beta::Types::Scalars::DateTime, null: false

        def user
          Beta::RecordLoader.for(User).load(object.user_id)
        end

        def work
          Beta::RecordLoader.for(Work).load(object.work_id)
        end

        def records
          Beta::ForeignKeyLoader.for(EpisodeRecord, :multiple_episode_record_id).load([object.id])
        end
      end
    end
  end
end
