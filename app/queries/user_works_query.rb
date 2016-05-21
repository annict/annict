class UserWorksQuery
  def initialize(user)
    @user = user
  end

  def watching
    Work.joins(:latest_statuses).merge(@user.latest_statuses.watching)
  end

  def wanna_watch_and_watching
    Work.joins(:latest_statuses).merge(@user.latest_statuses.wanna_watch_and_watching)
  end

  def desiring_to_watch
    Work.joins(:latest_statuses).merge(@user.latest_statuses.desiring_to_watch)
  end

  def on_hold
    Work.joins(:latest_statuses).merge(@user.latest_statuses.on_hold)
  end

  def unknown
    Work.where.not(id: @user.latest_statuses.pluck(:work_id))
  end

  def on(status_kind)
    Work.joins(:latest_statuses).merge(@user.latest_statuses.with_kind(status_kind))
  end

  def all
    Work.joins(:latest_statuses).merge(@user.latest_statuses)
  end
end
