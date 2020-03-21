# frozen_string_literal: true

module Api
  module Internal
    class UserSlotsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def index
        @user = current_user
        @slots = UserSlotsQuery.new(
          @user,
          Slot.only_kept.with_works(@user.works_on(:wanna_watch, :watching).only_kept),
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
