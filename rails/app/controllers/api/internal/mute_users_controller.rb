# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class MuteUsersController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        user = User.only_kept.find(params[:user_id])

        if current_user.mute(user)
          render(json: {flash: {type: :notice, message: t("messages._components.mute_user_button.the_user_has_been_muted")}}, status: 201)
        else
          head :bad_request
        end
      end

      def destroy
        mute_user = current_user.mute_users.find_by!(muted_user_id: params[:user_id])
        mute_user.destroy
        render(json: {flash: {type: :notice, message: t("messages._components.mute_user_button.the_user_has_been_unmuted")}}, status: 200)
      end
    end
  end
end
