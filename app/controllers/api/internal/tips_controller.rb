# frozen_string_literal: true

module Api
  module Internal
    class TipsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(close)

      def close
        slug = params[:slug]
        UserTipsService.new(current_user).finish!(slug)
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:tips, :close, el: slug, ds: "internal_api")
        head 200
      end
    end
  end
end
