# frozen_string_literal: true

module Api
  module Internal
    class FollowingController < Api::Internal::ApplicationController
      def index
        return render(json: []) unless user_signed_in?

        following = current_user.followings.only_kept.pluck(:id).map do |user_id|
          {
            user_id: user_id
          }
        end

        render json: following
      end
    end
  end
end
