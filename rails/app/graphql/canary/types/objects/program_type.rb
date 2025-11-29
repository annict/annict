# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ProgramType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer, null: false
        field :channel, Canary::Types::Objects::ChannelType, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :started_at, Canary::Types::Scalars::DateTime, null: true
        field :vod_title_code, String, null: false
        field :vod_title_name, String, null: false
        field :vod_title_url, String, null: false
        field :rebroadcast, Boolean, null: false
        field :viewer_selected, Boolean, null: false

        field :slots, Canary::Types::Objects::SlotType.connection_type, null: false, resolver: Canary::Resolvers::SlotsOnProgram do
          argument :viewer_untracked, Boolean, required: false
          argument :order_by, Canary::Types::InputObjects::SlotOrder, required: false
        end

        def channel
          RecordLoader.for(Channel).load(object.channel_id)
        end

        def work
          RecordLoader.for(Work).load(object.work_id)
        end

        def viewer_selected
          RecordLoader
            .for(LibraryEntry, column: :program_id, where: {user_id: context[:viewer].id})
            .load(object.id).then do |le|
              !le.nil?
            end
        end
      end
    end
  end
end
