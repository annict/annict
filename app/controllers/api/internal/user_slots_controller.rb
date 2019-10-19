# frozen_string_literal: true

module Api
  module Internal
    class UserSlotsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index
        @user = current_user
        @slots = current_user.slots.unwatched(params[:page], params[:sort])
      end
    end
  end
end
