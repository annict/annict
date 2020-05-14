# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ActivityType < Canary::Types::Objects::Base
        graphql_name "Activity"

        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :resource_type, Canary::Types::Enums::ActivityType, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :resource, Canary::Types::Unions::ActivityItem, null: false

        def resource_type
          object.resource_type.upcase
        end

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def resource
          case object.resource_type.to_sym
          when :episode_record
            RecordLoader.for(EpisodeRecord).load(object.episode_record_id)
          when :status
            RecordLoader.for(Status).load(object.status_id)
          when :work_record
            RecordLoader.for(WorkRecord).load(object.work_record_id)
          end
        end
      end
    end
  end
end
