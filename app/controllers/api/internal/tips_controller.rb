# frozen_string_literal: true

module Api
  module Internal
    class TipsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: [:finish]

      def finish(partial_name)
        UserTipsService.new(current_user).finish!(partial_name)
        head 200
      end
    end
  end
end
