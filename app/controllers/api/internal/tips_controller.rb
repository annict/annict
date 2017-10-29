# frozen_string_literal: true

module Api
  module Internal
    class TipsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(close)

      def close(slug, page_category)
        UserTipsService.new(current_user).finish!(slug)
        ga_client.page_category = page_category
        ga_client.events.create(:tips, :close, ev: slug)
        keen_client.publish(
          "close_tips",
          user: current_user,
          page_category: page_category,
          via: "internal_api",
          slug: slug
        )
        head 200
      end
    end
  end
end
