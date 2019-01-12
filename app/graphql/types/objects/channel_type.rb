# frozen_string_literal: true

module Types
  module Objects
    class ChannelType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :programs, Types::Objects::ProgramType.connection_type, null: true
      field :channel_group, Types::Objects::ChannelGroupType, null: false
      field :sc_chid, Integer, null: false
      field :name, String, null: false
      field :published, Boolean, null: false

      def programs
        ForeignKeyLoader.for(Program, :channel_id).load([object.id])
      end

      def channel_group
        RecordLoader.for(ChannelGroup).load(object.channel_group_id)
      end

      def sc_chid
        object.sc_chid
      end
    end
  end
end
