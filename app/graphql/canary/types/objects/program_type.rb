# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ProgramType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: false
        field :channel, Canary::Types::Objects::ChannelType, null: false
        field :anime, Canary::Types::Objects::AnimeType, null: false
        field :started_at, Canary::Types::Scalars::DateTime, null: true
        field :vod_title_code, String, null: false
        field :vod_title_name, String, null: false
        field :vod_title_url, String, null: false
        field :rebroadcast, Boolean, null: false
        field :viewer_did_check, Boolean, null: false

        field :slots, Canary::Types::Objects::SlotType.connection_type, null: true do
          argument :order_by, Canary::Types::InputObjects::SlotOrder, required: false
        end

        def channel
          RecordLoader.for(Channel).load(object.channel_id)
        end

        def anime
          RecordLoader.for(Work).load(object.work_id)
        end

        def slots(order_by: nil)
          SlotsQuery.new(
            object.slots.only_kept,
            order: Canary::OrderProperty.build(order_by)
          ).call
        end

        def viewer_did_check
          RecordLoader.
            for(LibraryEntry, column: :program_id, where: { user_id: context[:viewer].id }).
            load(object.id).then do |le|
              !le.nil?
            end
        end
      end
    end
  end
end
