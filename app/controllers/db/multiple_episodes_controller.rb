class Db::MultipleEpisodesController < Db::ApplicationController
  permits :body

  before_action :authenticate_user!
  before_action :load_work, only: [:new, :create]

  def new
    @me_form = DB::MultipleEpisodesForm.new
    authorize @me_form, :new?
  end

  def create(db_multiple_episodes_form)
    @me_form = DB::MultipleEpisodesForm.new(db_multiple_episodes_form)
    authorize @me_form, :create?

    @me_form.work = @work

    if @me_form.save_and_create_db_activity(current_user, "multiple_episodes.create")
      flash[:notice] = "エピソードを登録しました"
      redirect_to db_work_episodes_path(@work)
    else
      render :new
    end
  end

  private

  def load_work
    @work = Work.find(params[:work_id])
  end
end
