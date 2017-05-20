# frozen_string_literal: true

ObjectTypes::Episode = GraphQL::ObjectType.define do
  name "Episode"
  description "An episode of a work"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  connection :records, ObjectTypes::Record.connection_type do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(Checkin, :episode_id).load([obj.id])
    }
  end

  field :number, !types.String
  field :sort_number, !types.Int

  field :title, !types.String

  field :records_count, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.checkins_count
    }
  end

  field :record_comments_count, !types.Int

  field :work, !ObjectTypes::Work do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Work).load(obj.work_id)
    }
  end

  field :prev_episode, ObjectTypes::Episode do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.prev_episode_id)
    }
  end

  field :next_episode, ObjectTypes::Episode do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.next_episode_id)
    }
  end
end
