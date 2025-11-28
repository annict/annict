# typed: false
# frozen_string_literal: true

module Beta
  module Types
    module Objects
      class ChannelGroupType < Beta::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :annict_id, Integer, null: false
        field :channels, Beta::Types::Objects::ChannelType.connection_type, null: true
        field :name, String, null: false
        field :sort_number, Integer, null: false

        def channels
          Beta::ForeignKeyLoader.for(Channel, :channel_group_id).load([object.id])
        end

        def sort_number
          object.sort_number
        end
      end
    end
  end
end
