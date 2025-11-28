# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ChannelType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: false
        field :slots, Canary::Types::Objects::SlotType.connection_type, null: true
        field :channel_group, Canary::Types::Objects::ChannelGroupType, null: false
        field :sc_chid, Integer,
          null: false,
          description: "しょぼいカレンダーのチャンネルID"
        field :name, String, null: false

        def slots
          ForeignKeyLoader.for(Program, :channel_id).load([object.id])
        end

        def channel_group
          RecordLoader.for(ChannelGroup).load(object.channel_group_id)
        end
      end
    end
  end
end
