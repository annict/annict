# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class ReceptionsController < Api::Internal::ApplicationController
      def create
        return head(:unauthorized) unless user_signed_in?

        @channel = Channel.only_kept.find(params[:channel_id])
        current_user.receive(@channel)

        head 201
      end

      def destroy
        return head(:unauthorized) unless user_signed_in?

        @channel = Channel.only_kept.find(params[:channel_id])
        current_user.unreceive(@channel)

        head 200
      end
    end
  end
end
