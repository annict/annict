class ShotsController < ApplicationController
  layout 'application_no_navbar'

  def show(username)
    @user = User.find_by(username: username)
    works1 = @user.works.watching_with_season
    works2 = @user.works.watching.where(season_id: nil).select(:id, :title)
    @seasons = works1.group_by(&:season_name)
    @seasons = @seasons.merge('その他' => works2.to_a) if works2.present?
  end
end
