# frozen_string_literal: true

module Beta
  module Edges
    class ActivityEdge < GraphQL::Types::Relay::BaseEdge
      node_type Beta::Types::Objects::ActivityType

      field :annict_id, Integer, null: false
      field :user, Beta::Types::Objects::UserType, null: false
      field :action, Beta::Types::Enums::ActivityAction, null: false
      field :node, Beta::Types::Unions::ActivityItem, null: true

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

        case activity.itemable_type
        when "Status"
          Beta::RecordLoader.for(Status).load(activity.itemable_id)
        when "Record"
          Beta::RecordLoader.for(Record).load(activity.itemable_id).then do |record|
            if record.episode_record?
              Beta::RecordLoader.for(EpisodeRecord).load(record.recordable_id)
            else
              Beta::RecordLoader.for(WorkRecord).load(record.recordable_id)
            end
          end
        end
      end
    end
  end
end
