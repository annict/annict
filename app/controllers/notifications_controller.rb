# frozen_string_literal: true
# == Schema Information
#
# Table name: notifications
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  action_user_id :integer          not null
#  trackable_id   :integer          not null
#  trackable_type :string           not null
#  action         :string           not null
#  read           :boolean          default(FALSE), not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  index_notifications_on_read                             (read)
#  index_notifications_on_trackable_id_and_trackable_type  (trackable_id,trackable_type)
#

class NotificationsController < ApplicationController
  before_action :authenticate_user!

  def index(page: nil)
    @notifications = current_user.
      notifications.
      includes(:action_user).
      order(created_at: :desc).
      page(page)

    current_user.read_notifications! if 0 < current_user.notifications_count

    render layout: "v1/application"
  end
end
