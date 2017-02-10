# frozen_string_literal: true

module Api
  module Internal
    class ReceptionsController < Api::Internal::ApplicationController
      before_action :authenticate_user!
      before_action :load_channel, only: %i(create destroy)

      def create
        current_user.receive(@channel)
        head 200
      end

      def destroy
        current_user.unreceive(@channel)
        head 200
      end

      private

      def load_channel
        channel_id = params[:id] || params[:channel_id]
        @channel = Channel.published.find(channel_id)
      end
    end
  end
end
