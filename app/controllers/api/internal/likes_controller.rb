# frozen_string_literal: true

module Api
  module Internal
    class LikesController < Api::Internal::ApplicationController
      def index
        return render(json: []) unless user_signed_in?

        columns = %i[likeable_type likeable_id]
        likes = current_user.likes.pluck(*columns).map do |like|
          columns.zip(like).to_h
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
