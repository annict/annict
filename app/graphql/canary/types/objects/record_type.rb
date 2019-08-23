# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class RecordType < Canary::Types::Objects::Base
        implements GraphQL::Relay::Node.interface

        global_id_field :id

        field :annict_id, Integer, null: false
        field :user, Canary::Types::Objects::UserType, null: false
        field :work, Canary::Types::Objects::WorkType, null: false
        field :page_views_count, Integer, null: false

        def user
          Canary::RecordLoader.for(User).load(object.user_id)
        end

        def work
          Canary::RecordLoader.for(Work).load(object.work_id)
        end

        def page_views_count
          object.impressions_count
        end
      end
    end
  end
end
