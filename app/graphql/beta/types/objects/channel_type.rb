# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class ChannelType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :programs, Beta::Types::Objects::ProgramType.connection_type, null: true
        field :channel_group, Beta::Types::Objects::ChannelGroupType, null: false
        field :sc_chid, Integer, null: false
        field :name, String, null: false
        field :published, Boolean, null: false

        def programs
          Beta::ForeignKeyLoader.for(Program, :channel_id).load([object.id])
        end

        def channel_group
          Beta::RecordLoader.for(ChannelGroup).load(object.channel_group_id)
        end

        def sc_chid
          object.sc_chid
        end

        def published
          object.not_deleted?
        end
      end
    end
  end
end
