# frozen_string_literal: true

ObjectTypes::Channel = GraphQL::ObjectType.define do
  name "Channel"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  connection :programs, ObjectTypes::Program.connection_type do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(Program, :channel_id).load([obj.id])
    }
  end

  field :channelGroup, !ObjectTypes::ChannelGroup do
    resolve ->(obj, _args, _ctx) {
      RecordLoader.for(ChannelGroup).load(obj.channel_group_id)
    }
  end

  field :scChid, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.sc_chid
    }
  end

  field :name, !types.String
  field :published, !types.Boolean
end
