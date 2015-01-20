class UserWorksQuery
  def initialize(user)
    @user = user
  end

  def watching
    Work.joins(:statuses).merge(@user.statuses.watching)
  end

  def wanna_watch_and_watching
    Work.joins(:statuses).merge(@user.statuses.wanna_watch_and_watching)
  end

  def unknown
    Work.where.not(id: @user.statuses.latest.pluck(:work_id))
  end

  def on(status_kind)
    Work.joins(:statuses).merge(@user.statuses.latest.with_kind(status_kind))
  end

  def watching_with_season
    columns = ['works.id', 'works.title', 'seasons.name as season_name']
    watching.where.not(season_id: nil).order('seasons.id DESC')
      .joins(:season).select(columns)
  end
end
