# frozen_string_literal: true

module API
  module Internal
    class ReceptionsController < API::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        @channel = Channel.without_deleted.find(channel_id)
        current_user.receive(@channel)
        head 200
      end

      def destroy
        @channel = Channel.without_deleted.find(channel_id)
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
