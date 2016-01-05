# == Schema Information
#
# Table name: statuses
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  work_id     :integer          not null
#  kind        :integer          not null
#  latest      :boolean          default(FALSE), not null
#  likes_count :integer          default(0), not null
#  created_at  :datetime
#  updated_at  :datetime
#
# Indexes
#
#  statuses_user_id_idx  (user_id)
#  statuses_work_id_idx  (work_id)
#

class StatusesController < ApplicationController
  before_action :authenticate_user!
  before_action :set_work


  def select(status_kind)
    if Status.kind.values.include?(status_kind)
      status = current_user.statuses.new(work_id: @work.id, kind: status_kind)

      if status.save
        keen_client.statuses.create(status)
        render status: 200, nothing: true
      end
    elsif status_kind == 'no_select'
      status = current_user.statuses.find_by(work_id: @work.id, latest: true)
      status.toggle!(:latest) if status.present?
      render status: 200, nothing: true
    end
  end
end
