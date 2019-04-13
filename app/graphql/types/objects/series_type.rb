# frozen_string_literal: true

module Types
  module Objects
    class SeriesType < Types::Objects::Base
      implements GraphQL::Relay::Node.interface

      global_id_field :id

      field :annict_id, Integer, null: false
      field :name, String, null: false
      field :name_ro, String, null: false
      field :name_en, String, null: false

      field :works, Connections::SeriesWorkConnection, null: true, connection: true do
        argument :order_by, Types::InputObjects::SeriesWorkOrder, required: false
      end

      def works(order_by: nil)
        SearchSeriesWorksQuery.new(
          object.series_works,
          order_by: order_by
        ).call
      end
    end
  end
end
