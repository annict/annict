# frozen_string_literal: true

module V3
  class NotificationsController < V3::ApplicationController
    before_action :authenticate_user!

    def index
      @notifications = current_user
        .notifications
        .includes(:action_user)
        .order(created_at: :desc)
        .page(params[:page])

      current_user.read_notifications! if current_user.notifications_count.positive?
    end
  end
end
