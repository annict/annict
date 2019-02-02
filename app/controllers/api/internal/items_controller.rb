# frozen_string_literal: true

module Api
  module Internal
    class ItemsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(create)

      def create(resource_type, resource_id, asin, page_category)
        CreateItemJob.perform_later(current_user.id, resource_type, resource_id, asin)
        ga_client.page_category = page_category
        ga_client.events.create(:items, :create, el: resource_type, ev: resource_id, ds: "internal_api")
        keen_client.publish(:item_create, via: "internal_api")
        head 201
      end
    end
  end
end
