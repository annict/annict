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

  field :action, !EnumTypes::ActivityAction do
    resolve ->(obj, _args, _ctx) {
      activity = obj.node

      case activity.action
      when "create_status" then "CREATE"
      when "create_record" then "CREATE"
      when "create_review" then "CREATE"
      when "create_multiple_records" then "CREATE"
      end
    }
  end

  field :node, UnionTypes::ActivityItem do
    resolve ->(obj, _args, _ctx) {
      activity = obj.node

      case activity.trackable_type
      when "Status"
        RecordLoader.for(Status).load(activity.trackable_id)
      when "Record"
        RecordLoader.for(Record).load(activity.trackable_id)
      when "WorkRecord"
        RecordLoader.for(WorkRecord).load(activity.trackable_id)
      when "MultipleRecord"
        RecordLoader.for(MultipleRecord).load(activity.trackable_id)
      end
    }
  end
end
