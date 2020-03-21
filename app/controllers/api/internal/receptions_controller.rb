# frozen_string_literal: true

module Api
  module Internal
    class ReceptionsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        @channel = Channel.only_kept.find(channel_id)
        current_user.receive(@channel)
        head 200
      end

      def destroy
        @channel = Channel.only_kept.find(channel_id)
        current_user.unreceive(@channel)
        head 200
      end

      private

      def channel_id
        params[:id] || params[:channel_id]
      end
    end
  end
end
