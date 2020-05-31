# frozen_string_literal: true

module Api
  module Internal
    class UsersController < Api::Internal::ApplicationController
      def show
        return render(json: {}) unless user_signed_in?

        render json: {
          hide_record_body: current_user.setting.hide_record_body
        }
      end
    end
  end
end
