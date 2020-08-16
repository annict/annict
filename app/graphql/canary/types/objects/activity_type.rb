# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class ActivityType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :database_id, Integer, null: false
        field :itemable_type, Canary::Types::Enums::ActivityItemableType, null: false
        field :created_at, Canary::Types::Scalars::DateTime, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :itemable, Canary::Types::Unions::ActivityItemable, null: false

        def itemable_type
          value = object.itemable_type.underscore.upcase

          case value
          when "WORK_RECORD"
            "ANIME_RECORD"
          else
            value
          end
        end

        def user
          RecordLoader.for(User).load(object.user_id)
        end

        def itemable
          case object.itemable_type
          when "EpisodeRecord"
            RecordLoader.for(EpisodeRecord).load(object.itemable_id)
          when "Status"
            RecordLoader.for(Status).load(object.itemable_id)
          when "WorkRecord"
            RecordLoader.for(WorkRecord).load(object.itemable_id)
          end
        end
      end
    end
  end
end
