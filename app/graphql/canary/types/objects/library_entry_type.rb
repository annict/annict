# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class LibraryEntryType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :tracked_episodes_count, Int, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :anime, Canary::Types::Objects::AnimeType, null: false
        field :status, Canary::Types::Objects::StatusType, null: true
        field :program, Canary::Types::Objects::ProgramType, null: true

        def tracked_episodes_count
          object.watched_episode_ids.size
        end

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def anime
          RecordLoader.for(Work).load(object.work_id)
        end

        def status
          RecordLoader.for(Status).load(object.status_id)
        end

        def program
          RecordLoader.for(Program).load(object.program_id)
        end
      end
    end
  end
end
