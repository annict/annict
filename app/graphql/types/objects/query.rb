# frozen_string_literal: true

module Types
  module Objects
    class Query < Types::Objects::Base
      field :node, field: GraphQL::Relay::Node.field
      field :nodes, field: GraphQL::Relay::Node.plural_field
      field :viewer, Types::Objects::UserType, null: true

      field :user, Types::Objects::UserType, null: true do
        argument :username, String, required: true
      end

      field :search_works, Types::Objects::WorkType.connection_type, null: true do
        argument :annict_ids, [Integer], required: false
        argument :seasons, [String], required: false
        argument :titles, [String], required: false
        argument :order_by, Types::InputObjects::WorkOrder, required: false
      end

      def viewer
        context[:viewer]
      end

      def user(username:)
        User.published.find_by(username: username)
      end

      def search_works(annict_ids: nil, seasons: nil, titles: nil, order_by: nil)
        SearchWorksQuery.new(
          annict_ids: annict_ids,
          seasons: seasons,
          titles: titles,
          order_by: order_by
        ).call
      end
    end
  end
end
