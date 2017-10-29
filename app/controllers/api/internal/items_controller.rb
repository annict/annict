# frozen_string_literal: true

module Api
  module Internal
    class ItemsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(create)

      def create(resource_type, resource_id, asin, page_category)
        CreateItemJob.perform_later(current_user.id, resource_type, resource_id, asin)
        ga_client.page_category = page_category
        ga_client.events.create(:items, :create)
        keen_client.publish(
          "create_items",
          user: current_user,
          page_category: page_category,
          via: "internal_api",
          resource_type: resource_type
        )
        head 201
      end
    end
  end
end
