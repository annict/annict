# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class SeriesType < Canary::Types::Objects::Base
        implements GraphQL::Types::Relay::Node

        global_id_field :id

        field :database_id, Integer,
          null: false

        field :name, String,
          null: false

        field :name_en, String,
          null: false

        field :work_list, Canary::Connections::SeriesWorkConnection, null: true, connection: true do
          argument :order_by, Canary::Types::InputObjects::SeriesWorkOrder, required: false
        end

        def work_list(order_by: nil)
          SearchSeriesWorksQuery.new(
            object.series_works,
            order_by: order_by
          ).call
        end
      end
    end
  end
end
