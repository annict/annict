# frozen_string_literal: true
# == Schema Information
#
# Table name: statuses
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  work_id              :integer          not null
#  kind                 :integer          not null
#  likes_count          :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  oauth_application_id :integer
#
# Indexes
#
#  index_statuses_on_oauth_application_id  (oauth_application_id)
#  statuses_user_id_idx                    (user_id)
#  statuses_work_id_idx                    (work_id)
#

class StatusesController < ApplicationController
  before_action :authenticate_user!
  before_action :set_work

  def select(status_kind)
    status = StatusService.new(current_user, @work, ga_client)
    render(status: 200, nothing: true) if status.change(status_kind)
  end
end
