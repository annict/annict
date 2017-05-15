# frozen_string_literal: true

Types::EpisodeType = GraphQL::ObjectType.define do
  name "Episode"
  description "An episode of a work"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :number, !types.String
  field :sort_number, !types.Int

  field :title, !types.String

  field :records_count, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.checkins_count
    }
  end

  field :record_comments_count, !types.Int

  field :work, !Types::WorkType do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Work).load(obj.work_id)
    }
  end

  field :prev_episode, Types::EpisodeType do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.prev_episode_id)
    }
  end

  field :next_episode, Types::EpisodeType do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.next_episode_id)
    }
  end
end
