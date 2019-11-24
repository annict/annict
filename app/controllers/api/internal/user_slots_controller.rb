# frozen_string_literal: true

module Api
  module Internal
    class UserSlotsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index
        @user = current_user
        @slots = UserSlotsQuery.new(
          @user,
          Slot.without_deleted,
          status_kinds: %i(wanna_watch watching),
          watched: false,
          order: order_property(params[:sort])
        ).call.page(params[:page])
      end

      private

      def order_property(sort_type)
        case sort_type
        when "started_at_asc"
          OrderProperty.new(:started_at, :asc)
        else
          OrderProperty.new(:started_at, :desc)
        end
      end
    end
  end
end
