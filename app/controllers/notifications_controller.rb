class NotificationsController < ApplicationController
  before_filter :authenticate_user!

  def index(page: nil)
    @notifications = current_user
                       .notifications
                       .includes(:action_user)
                       .order(created_at: :desc)
                       .page(page)

    current_user.read_notifications! if 0 < current_user.notifications_count
  end
end
