# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ActivityGroupType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: false
        field :itemable_type, Canary::Types::Enums::ActivityItemableType, null: false
        field :single, Boolean, null: false
        field :activities_count, Integer, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :activities, Canary::Types::Objects::ActivityType.connection_type, null: false

        def itemable_type
          value = object.itemable_type.underscore.upcase

          case value
          when "EPISODE_RECORD", "WORK_RECORD"
            "RECORD"
          else
            value
          end
        end

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def activities
          Canary::AssociationLoader.for(ActivityGroup, %i[ordered_activities]).load(object)
        end
      end
    end
  end
end
