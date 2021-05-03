# frozen_string_literal: true

module Api
  module Internal
    class LikesController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i[unlike]

      def index
        return render(json: []) unless user_signed_in?

        likes = current_user.likes.pluck(:recipient_id, :recipient_type).map { |(recipient_id, recipient_type)|
          {
            recipient_type: recipient_type,
            recipient_id: recipient_id
          }
        }

        render json: likes
      end

      def create
        return head(:unauthorized) unless user_signed_in?

        recipient = params[:recipient_type].constantize.find(params[:recipient_id])

        AddReactionRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(reactable: recipient, content: "HEART")

        head 200
      end

      def unlike
        return head(:unauthorized) unless user_signed_in?

        recipient = params[:recipient_type].constantize.find(params[:recipient_id])

        RemoveReactionRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(reactable: recipient, content: "HEART")

        head 200
      end
    end
  end
end
