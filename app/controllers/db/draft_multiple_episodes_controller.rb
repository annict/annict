class Db::DraftMultipleEpisodesController < Db::ApplicationController
  permits :body, edit_request_attributes: [:id, :title, :body]

  before_action :authenticate_user!
  before_action :set_work, only: [:new, :create, :edit, :update]

  def new(work_id)
    @draft_multiple_episode = @work.draft_multiple_episodes.new
    @draft_multiple_episode.build_edit_request
  end

  def create(draft_multiple_episode)
    @draft_multiple_episode = @work.draft_multiple_episodes.new(draft_multiple_episode)
    @draft_multiple_episode.edit_request.user = current_user

    if @draft_multiple_episode.save
      flash[:notice] = "エピソードの編集リクエストを作成しました"
      redirect_to db_edit_request_path(@draft_multiple_episode.edit_request)
    else
      render :new
    end
  end

  def edit(id)
    @draft_multiple_episode = @work.draft_multiple_episodes.find(id)
    authorize @draft_multiple_episode, :edit?
  end

  def update(id, draft_multiple_episode)
    @draft_multiple_episode = @work.draft_multiple_episodes.find(id)
    authorize @draft_multiple_episode, :update?

    if @draft_multiple_episode.update(draft_multiple_episode)
      flash[:notice] = "作品の編集リクエストを更新しました"
      redirect_to db_edit_request_path(@draft_multiple_episode.edit_request)
    else
      render :edit
    end
  end

  private

  def set_work
    @work = Work.find(params[:work_id])
  end
end
