class Api::CheckinsController < Api::ApplicationController
  permits :comment, :shared_twitter, :shared_facebook, :spoil

  before_filter :authenticate_user!
  before_filter :set_work, only: [:create, :create_all, :update]
  before_filter :set_episode, only: [:create]
  before_filter :set_checkin, only: [:update]

  def create(checkin)
    @checkin = @episode.checkins.new(checkin)
    @checkin.user = current_user
    @checkin.work = @work

    if @checkin.save
      render status: 201
    else
      render status: 400, json: { message: @checkin.errors.first }
    end
  end

  def create_all(episode_ids)
    episodes = Episode.where(id: episode_ids).order(:sort_number)

    # 一括チェックインによって「Twitter/Facebookにシェアする」のチェックが外れないようにする
    Checkin.skip_callback(:save, :after, :update_share_checkin_status)

    episodes.each do |episode|
      episode.checkins.create(user: current_user, work: @work)
    end
  end

  def update(checkin)
    if @checkin.update_attributes(checkin)
      redirect_to work_episode_checkin_path(@work, @episode, @checkin), notice: t('checkins.updated')
    else
      render :edit
    end
  end

  private

  def set_work
    @work = Work.find(params[:work_id])
  end

  def set_checkin
    @checkin = @episode.checkins.find(params[:id])
  end
end
