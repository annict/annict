# frozen_string_literal: true

module Types
  module Objects
    class StatusType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :user, Types::Objects::UserType, null: false
      field :work, Types::Objects::WorkType, null: false
      field :state, Types::Enums::StatusState, null: false
      field :likes_count, Integer, null: false
      field :created_at, Types::Scalars::DateTime, null: false

      def user
        RecordLoader.for(User).load(object.user_id)
      end

      def work
        RecordLoader.for(Work).load(object.work_id)
      end

      def state
        object.kind.upcase
      end
    end
  end
end
