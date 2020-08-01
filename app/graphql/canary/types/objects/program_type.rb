# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ProgramType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: false
        field :channel, Canary::Types::Objects::ChannelType, null: false
        field :work, Canary::Types::Objects::AnimeType, null: false
        field :started_at, Canary::Types::Scalars::DateTime, null: false
        field :vod_title_code, String, null: false
        field :vod_title_name, String, null: false
        field :vod_title_url, String, null: false
        field :rebroadcast, Boolean, null: false

        field :slots, Canary::Types::Objects::SlotType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::SlotOrder, required: false
        end

        def channel
          RecordLoader.for(Channel).load(object.channel_id)
        end

        def work
          RecordLoader.for(Anime).load(object.work_id)
        end

        def slots(order_by: nil)
          SlotsQuery.new(
            object.slots.only_kept,
            order: build_order(order_by)
          ).call
        end
      end
    end
  end
end
