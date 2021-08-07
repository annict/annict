# frozen_string_literal: true

module Api
  module Internal
    class UnlikesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        recipient = params[:recipient_type].constantize.find(params[:recipient_id])
        Creators::UnlikeCreator.new(user: current_user, likeable: recipient).call

        head 201
      end
    end
  end
end
