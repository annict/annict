# frozen_string_literal: true

class NotificationsController < ApplicationV6Controller
  before_action :authenticate_user!

  def index
    @notifications = current_user
      .notifications
      .includes(:action_user)
      .order(created_at: :desc)
      .page(params[:page])
      .without_count

    current_user.read_notifications! if current_user.notifications_count.positive?
  end
end
