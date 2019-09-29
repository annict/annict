# frozen_string_literal: true

module Canary
  module Edges
    class ActivityEdge < GraphQL::Types::Relay::BaseEdge
      node_type Canary::Types::Objects::ActivityType

      field :annict_id, Integer, null: false
      field :user, Canary::Types::Objects::UserType, null: false
      field :action, Canary::Types::Enums::ActivityAction, null: false
      field :node, Canary::Types::Unions::ActivityItem, null: true

      def annict_id
        activity = object.node
        activity.id
      end

      def user
        activity = object.node
        RecordLoader.for(User).load(activity.user_id)
      end

      def action
        activity = object.node

        case activity.action
        when "create_status" then "CREATE"
        when "create_episode_record" then "CREATE"
        when "create_work_record" then "CREATE"
        when "create_multiple_episode_records" then "CREATE"
        end
      end

      def node
        activity = object.node

        case activity.trackable_type
        when "Status"
          RecordLoader.for(Status).load(activity.trackable_id)
        when "EpisodeRecord"
          RecordLoader.for(EpisodeRecord).load(activity.trackable_id)
        when "WorkRecord"
          RecordLoader.for(WorkRecord).load(activity.trackable_id)
        when "MultipleEpisodeRecord"
          RecordLoader.for(MultipleEpisodeRecord).load(activity.trackable_id)
        end
      end
    end
  end
end
