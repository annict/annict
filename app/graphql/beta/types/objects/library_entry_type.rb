# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class LibraryEntryType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :user, Beta::Types::Objects::UserType, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :status, Beta::Types::Objects::StatusType, null: true
        field :next_episode, Beta::Types::Objects::EpisodeType, null: true
        field :next_program, Beta::Types::Objects::ProgramType, null: true
        field :note, String, null: false

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

        def next_program
          RecordLoader.for(Slot).load(object.next_slot_id)
        end
      end
    end
  end
end
