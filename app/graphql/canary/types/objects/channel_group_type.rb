# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ChannelGroupType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: false
        field :channels, Canary::Types::Objects::ChannelType.connection_type, null: true
        field :name, String, null: false
        field :sort_number, Integer,
          null: false,
          description: "ソート番号"

        def channels
          ForeignKeyLoader.for(Channel, :channel_group_id).load([object.id])
        end
      end
    end
  end
end
