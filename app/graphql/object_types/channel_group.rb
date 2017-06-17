# frozen_string_literal: true

ObjectTypes::ChannelGroup = GraphQL::ObjectType.define do
  name "ChannelGroup"

  implements GraphQL::Relay::Node.interface

  global_id_field :id

  field :annictId, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.id
    }
  end

  connection :channels, ObjectTypes::Channel.connection_type do
    resolve ->(obj, _args, _ctx) {
      ForeignKeyLoader.for(Channel, :channel_group_id).load([obj.id])
    }
  end

  field :name, !types.String

  field :sortNumber, !types.Int do
    resolve ->(obj, _args, _ctx) {
      obj.sort_number
    }
  end
end
