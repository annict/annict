# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class MutedUsersController < Api::Internal::ApplicationController
      def index
        return render(json: []) unless user_signed_in?

        render(json: current_user.mute_users.pluck(:muted_user_id))
      end
    end
  end
end
