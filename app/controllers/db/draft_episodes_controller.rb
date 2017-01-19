class Db::DraftEpisodesController < Db::ApplicationController
  permits :number, :sort_number, :title, :prev_episode_id, :episode_id, :fetch_syobocal,
    :raw_number, :sc_count, edit_request_attributes: [:id, :title, :body]

  before_action :authenticate_user!
  before_action :set_work, only: [:new, :create, :edit, :update]

  def new(id)
    @episode = @work.episodes.find(id)
    attributes = @episode.attributes.slice(*Episode::DIFF_FIELDS.map(&:to_s))
    @draft_episode = @work.draft_episodes.new(attributes)
    authorize @draft_episode, :new?
    @draft_episode.build_edit_request
  end

  def create(draft_episode)
    @draft_episode = @work.draft_episodes.new(draft_episode)
    authorize @draft_episode, :create?
    @draft_episode.edit_request.user = current_user
    @episode = @work.episodes.find(draft_episode[:episode_id])
    @draft_episode.origin = @episode

    if @draft_episode.save
      flash[:notice] = "エピソードの編集リクエストを作成しました"
      redirect_to db_edit_request_path(@draft_episode.edit_request)
    else
      render :new
    end
  end

  def edit(id)
    @draft_episode = @work.draft_episodes.find(id)
  end

  def update(id, draft_episode)
    @draft_episode = @work.draft_episodes.find(id)

    if @draft_episode.update(draft_episode)
      flash[:notice] = "エピソードの編集リクエストを更新しました"
      redirect_to db_edit_request_path(@draft_episode.edit_request)
    else
      render :edit
    end
  end

  private

  def set_work
    @work = Work.find(params[:work_id])
  end
end
