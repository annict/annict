# frozen_string_literal: true

module Edges
  class ActivityEdge < GraphQL::Types::Relay::BaseEdge
    node_type Types::Objects::ActivityType

    field :annict_id, Integer, null: false
    field :user, Types::Objects::UserType, null: false
    field :action, Types::Enums::ActivityAction, null: false
    field :node, Types::Unions::ActivityItem, null: true

    def annict_id
      activity = object.node
      activity.id
    end

    def user
      activity = object.node
      RecordLoader.for(User).load(activity.user_id)
    end

    def action
      "CREATE"
    end

    def node
      activity = object.node

      case activity.trackable_type
      when "Status"
        RecordLoader.for(Status).load(activity.trackable_id)
      when "EpisodeRecord"
        RecordLoader.for(EpisodeRecord).load(activity.trackable_id)
      when "AnimeRecord"
        RecordLoader.for(AnimeRecord).load(activity.trackable_id)
      when "MultipleEpisodeRecord"
        RecordLoader.for(MultipleEpisodeRecord).load(activity.trackable_id)
      end
    end
  end
end
