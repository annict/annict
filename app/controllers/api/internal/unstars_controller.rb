# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class UnstarsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        starrable = params[:starrable_type].constantize.find(params[:starrable_id])
        current_user.unfavorite(starrable)

        render(json: {}, status: 201)
      end
    end
  end
end
