# frozen_string_literal: true

module Api
  module Internal
    class TipsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(close)

      def close(slug)
        UserTipsService.new(current_user).finish!(slug)
        keen_client.tips.close(current_user, slug)
        head 200
      end
    end
  end
end
