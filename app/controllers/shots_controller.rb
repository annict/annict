class ShotsController < ApplicationController
  layout 'application_no_navbar'

  def show(username)
    @user = User.find_by(username: username)
    columns = ['works.id', 'works.title', 'seasons.name as season_name']
    works1 = @user.watching_works.where.not(season_id: nil).order(released_at: :desc).joins(:season).select(columns)
    works2 = @user.watching_works.where(season_id: nil).select(:id, :title)
    @seasons = works1.group_by(&:season_name).merge('その他' => works2.to_a)
  end
end
