# frozen_string_literal: true

module Api
  module Internal
    class AmazonController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(search)

      def search(keyword, resource_id, resource_type, page: 1)
        @resource = resource_type.constantize.find(resource_id)
        client = Annict::Amazon::Client.new
        @res = client.items.search(keyword, item_page: page)
      end
    end
  end
end
