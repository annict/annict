# frozen_string_literal: true
# == Schema Information
#
# Table name: notifications
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  action_user_id :integer          not null
#  trackable_id   :integer          not null
#  trackable_type :string(510)      not null
#  action         :string(510)      not null
#  read           :boolean          default(FALSE), not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  notifications_action_user_id_idx  (action_user_id)
#  notifications_user_id_idx         (user_id)
#

class NotificationsController < ApplicationController
  before_action :authenticate_user!

  def index
    @notifications = current_user.
      notifications.
      includes(:action_user).
      order(created_at: :desc).
      page(params[:page])

    current_user.read_notifications! if current_user.notifications_count.positive?
  end
end
