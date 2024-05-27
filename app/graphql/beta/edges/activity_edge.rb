# typed: false
# frozen_string_literal: true

module Beta
  module Edges
    class ActivityEdge < GraphQL::Types::Relay::BaseEdge
      node_type Beta::Types::Objects::ActivityType

      field :annict_id, Integer, null: false
      field :user, Beta::Types::Objects::UserType, null: false
      field :action, Beta::Types::Enums::ActivityAction, null: false
      field :node, Beta::Types::Unions::ActivityItem, "Deprecated: Use `item` instead.", null: true, deprecation_reason: "Use `item` instead."
      field :item, Beta::Types::Unions::ActivityItem, null: true, resolver_method: :node

      def annict_id
        activity = object.node
        activity.id
      end

      def user
        activity = object.node
        Beta::RecordLoader.for(User).load(activity.user_id)
      end

      def action
        "CREATE"
      end

      def node
        activity = object.node

        case activity.trackable_type
        when "Status"
          Beta::RecordLoader.for(Status).load(activity.trackable_id)
        when "EpisodeRecord"
          Beta::RecordLoader.for(EpisodeRecord).load(activity.trackable_id)
        when "WorkRecord"
          Beta::RecordLoader.for(WorkRecord).load(activity.trackable_id)
        when "MultipleEpisodeRecord"
          Beta::RecordLoader.for(MultipleEpisodeRecord).load(activity.trackable_id)
        end
      end
    end
  end
end
