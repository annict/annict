# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class FollowingController < Api::Internal::ApplicationController
      def index
        return render(json: []) unless user_signed_in?

        render(json: current_user.followings.only_kept.pluck(:id))
      end
    end
  end
end
