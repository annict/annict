# frozen_string_literal: true

ObjectTypes::Program = GraphQL::ObjectType.define do
  name "Program"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  field :channel, !ObjectTypes::Channel do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Channel).load(obj.channel_id)
    }
  end

  field :episode, !ObjectTypes::Episode do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Episode).load(obj.episode_id)
    }
  end

  field :work, !ObjectTypes::Work do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(Work).load(obj.work_id)
    }
  end

  field :startedAt, !ScalarTypes::DateTime do
    resolve ->(obj, _args, _ctx) {
      obj.started_at
    }
  end

  field :scPid, types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.sc_pid
    }
  end

  field :state, !Types::Enum::ProgramState do
    resolve ->(obj, _args, _ctx) {
      obj.aasm_state.upcase
    }
  end

  field :rebroadcast, !types.Boolean
end
