# frozen_string_literal: true

module Api
  module Internal
    class LikesController < Api::Internal::ApplicationController
      def index
        return render(json: []) unless user_signed_in?

        likes = current_user.likes.preload(recipient: %i[record status]).map do |like|
          likeable = case like.recipient_type
          when "AnimeRecord", "EpisodeRecord", "WorkRecord"
            like.recipient.record
          else
            like.recipient
          end

          {
            recipient_type: likeable.class.name,
            recipient_id: likeable&.id
          }
        end

        render json: likes
      end

      def create
        return head(:unauthorized) unless user_signed_in?

        recipient = params[:recipient_type].constantize.find(params[:recipient_id])
        Creators::LikeCreator.new(user: current_user, likeable: recipient).call

        head 201
      end
    end
  end
end
