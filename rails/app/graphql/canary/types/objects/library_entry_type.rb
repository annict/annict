# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class LibraryEntryType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :tracked_episodes_count_in_current_status, Int, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :status, Canary::Types::Objects::StatusType, null: true
        field :program, Canary::Types::Objects::ProgramType, null: true

        def tracked_episodes_count_in_current_status
          object.watched_episode_ids.size
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

        def program
          RecordLoader.for(Program).load(object.program_id)
        end
      end
    end
  end
end
