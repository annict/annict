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

        field :anime_list, Canary::Connections::SeriesAnimeConnection, null: true, connection: true do
          argument :order_by, Canary::Types::InputObjects::SeriesAnimeOrder, required: false
        end

        def anime_list(order_by: nil)
          SearchSeriesWorksQuery.new(
            object.series_animes,
            order_by: order_by
          ).call
        end
      end
    end
  end
end
