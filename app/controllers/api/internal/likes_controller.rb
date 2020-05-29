# frozen_string_literal: true

module Api
  module Internal
    class LikesController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(unlike)

      def index
        return render(json: []) unless user_signed_in?

        likes = current_user.likes.pluck(:recipient_id, :recipient_type).map do |(recipient_id, recipient_type)|
          {
            recipient_type: recipient_type,
            recipient_id: recipient_id
          }
        end

        render json: likes
      end

      def create
        return head(:unauthorized) unless user_signed_in?

        recipient = params[:recipient_type].constantize.find(params[:recipient_id])
        current_user.like(recipient)
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:likes, :create, el: params[:recipient_type], ev: params[:recipient_id], ds: "internal_api")

        if params[:recipient_type] == "EpisodeRecord"
          EmailNotificationService.send_email(
            "liked_episode_record",
            recipient.user,
            current_user.id,
            params[:recipient_id]
          )
        end

        head 200
      end

      def unlike
        recipient = params[:recipient_type].constantize.find(params[:recipient_id])
        current_user.unlike(recipient)
        head 200
      end
    end
  end
end
