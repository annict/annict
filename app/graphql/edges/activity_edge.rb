# frozen_string_literal: true

Edges::ActivityEdge = ObjectTypes::Activity.define_edge do
  name "ActivityEdge"

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      activity = obj.node
      activity.id
    }
  end

  field :user, !ObjectTypes::User do
    resolve ->(obj, _args, _ctx) {
      activity = obj.node
      RecordLoader.for(User).load(activity.user_id)
    }
  end

  field :action, !Types::Enum::ActivityAction do
    resolve ->(obj, _args, _ctx) {
      activity = obj.node

      case activity.action
      when "create_status" then "CREATE"
      when "create_episode_record" then "CREATE"
      when "create_work_record" then "CREATE"
      when "create_multiple_episode_records" then "CREATE"
      end
    }
  end

  field :node, UnionTypes::ActivityItem do
    resolve ->(obj, _args, _ctx) {
      activity = obj.node

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
    }
  end
end
