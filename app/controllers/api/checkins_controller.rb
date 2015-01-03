class Api::CheckinsController < Api::ApplicationController
  before_filter :authenticate_user!
  before_filter :set_work

  def create_all(episode_ids)
    episodes = Episode.where(id: episode_ids).order(:sort_number)

    # 一括チェックインによって「Twitter/Facebookにシェアする」のチェックが外れないようにする
    Checkin.skip_callback(:save, :after, :update_share_checkin_status)

    episodes.each do |episode|
      episode.checkins.create(user: current_user, work: @work)
    end
  end

  private

  def set_work
    @work = Work.find(params[:work_id])
  end
end
