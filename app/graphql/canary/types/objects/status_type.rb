# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class StatusType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :kind, Canary::Types::Enums::StatusKind, null: false
        field :likes_count, Integer, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def work
          RecordLoader.for(Work).load(object.work_id)
        end

        def kind
          Status.kind_v2_to_v3(object.kind).upcase.to_s
        end
      end
    end
  end
end
