# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ActivityType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :resource_type, Canary::Types::Enums::ActivityResourceType, null: false
        field :action, Canary::Types::Enums::ActivityAction, null: false
        field :resources_count, Integer, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :resources, [Canary::Types::Unions::ActivityItem], null: true

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def resource_type
          case object.trackable_type
          when "EpisodeRecord" then "EPISODE_RECORD"
          when "Status"        then "STATUS"
          when "WorkRecord"    then "WORK_RECORD"
          end
        end

        def action
          "CREATE"
        end

        def resources
          Canary::AssociationLoader.for(Activity, %i(episode_records statuses work_records)).load(object)
        end
      end
    end
  end
end
