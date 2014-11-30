class ShotsController < ApplicationController
  layout 'application_no_navbar'

  def show(username)
    @user = User.find_by(username: username)
    columns = ['works.id', 'works.title', 'seasons.name as season_name']
    works1 = @user.works.watching.where.not(season_id: nil).order(released_at: :desc).joins(:season).select(columns)
    works2 = @user.works.watching.where(season_id: nil).select(:id, :title)
    @seasons = works1.group_by(&:season_name)
    @seasons = @seasons.merge('その他' => works2.to_a) if works2.present?
  end
end
