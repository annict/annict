# frozen_string_literal: true

module Api
  module Internal
    class FavoritesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        resource_type = params[:resource_type]
        resource_id = params[:resource_id]
        resource = resource_type.constantize.find(resource_id)
        current_user.favorite(resource)
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:favorites, :create, el: resource_type, ev: resource_id, ds: "internal_api")
        head 200
      end

      def unfavorite
        resource = params[:resource_type].constantize.find(params[:resource_id])
        current_user.unfavorite(resource)
        head 200
      end
    end
  end
end
