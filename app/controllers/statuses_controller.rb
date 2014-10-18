class StatusesController < ApplicationController
  before_filter :authenticate_user!
  before_filter :set_work


  def select(status_kind)
    if Status.kind.values.include?(status_kind)
      status = current_user.statuses.new(work_id: @work.id, kind: status_kind)

      if status.save
        render status: 200, nothing: true
      end
    elsif status_kind == 'no_select'
      status = current_user.statuses.find_by(work_id: @work.id, latest: true)
      status.toggle!(:latest) if status.present?
      render status: 200, nothing: true
    end
  end
end
