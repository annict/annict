# frozen_string_literal: true

ObjectTypes::Episode = GraphQL::ObjectType.define do
  name "Episode"
  description "An episode of a work"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  connection :records, ObjectTypes::Record.connection_type do
    argument :orderBy, Types::InputObjects::RecordOrder
    argument :hasComment, types.Boolean

    resolve Resolvers::Records.new
  end

  field :number, types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.raw_number
    }
  end

  field :numberText, types.String do
    resolve ->(obj, _args, _ctx) {
      obj.number
    }
  end

  field :sortNumber, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.sort_number
    }
  end

  field :title, types.String

  field :recordsCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.episode_records_count
    }
  end

  field :recordCommentsCount, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.episode_records_with_body_count
    }
  end

  field :work, !ObjectTypes::Work do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Work).load(obj.work_id)
    }
  end

  field :prevEpisode, ObjectTypes::Episode do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.prev_episode_id)
    }
  end

  field :nextEpisode, ObjectTypes::Episode do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.next_episode_id)
    }
  end

  field :viewerDidTrack, !types.Boolean do
    resolve ->(obj, _args, ctx) {
      ctx[:viewer].tracked?(obj)
    }
  end

  field :viewerRecordsCount, !types.Int do
    resolve ->(obj, _args, ctx) {
      ctx[:viewer].episode_records_count_in(obj)
    }
  end
end
