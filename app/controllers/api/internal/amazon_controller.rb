# frozen_string_literal: true

module Api
  module Internal
    class AmazonController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(search)

      def search
        @resource = params[:resource_type].constantize.find(params[:resource_id])
        client = Annict::Amazon::Client.new
        @res = client.items.search(params[:keyword], item_page: params[:page])
      end
    end
  end
end
