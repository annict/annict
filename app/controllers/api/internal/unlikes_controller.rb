# frozen_string_literal: true

module Api
  module Internal
    class UnlikesController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        likeable = params[:likeable_type].constantize.find(params[:likeable_id])
        Creators::UnlikeCreator.new(user: current_user, likeable: likeable).call

        head 201
      end
    end
  end
end
