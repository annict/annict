# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class StatusType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :user, Beta::Types::Objects::UserType, null: false
        field :work, Beta::Types::Objects::WorkType, null: false
        field :state, Beta::Types::Enums::StatusState, null: false
        field :likes_count, Integer, null: false
        field :created_at, Beta::Types::Scalars::DateTime, null: false

        def user
          Beta::RecordLoader.for(User).load(object.user_id)
        end

        def work
          Beta::RecordLoader.for(Anime).load(object.work_id)
        end

        def state
          object.kind.upcase
        end
      end
    end
  end
end
