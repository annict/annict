# frozen_string_literal: true

module Api
  module Internal
    class ItemsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(create)

      def create
        CreateItemJob.perform_later(current_user.id, params[:resource_type], params[:resource_id], params[:asin])
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:items, :create, el: params[:resource_type], ev: params[:resource_id], ds: "internal_api")
        keen_client.publish(:item_create, via: "internal_api")
        head 201
      end
    end
  end
end
