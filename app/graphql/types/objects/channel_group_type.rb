# frozen_string_literal: true

module Types
  module Objects
    class ChannelGroupType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :channels, Types::Objects::ChannelType.connection_type, null: true
      field :name, String, null: false
      field :sort_number, Integer, null: false

      def channels
        ForeignKeyLoader.for(Channel, :channel_group_id).load([object.id])
      end

      def sort_number
        object.sort_number
      end
    end
  end
end
