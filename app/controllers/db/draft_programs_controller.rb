class Db::DraftProgramsController < Db::ApplicationController
  permits :channel_id, :episode_id, :work_id, :started_at, :program_id,
          edit_request_attributes: [:id, :title, :body]

  before_action :set_work, only: [:new, :create, :edit, :update]

  def new(id: nil)
    @draft_program = if id.present?
      @program = @work.programs.find(id)
      attributes = @program.attributes.slice(*Program::DIFF_FIELDS.map(&:to_s))
      @work.draft_programs.new(attributes)
    else
      @work.draft_programs.new
    end
    @draft_program.build_edit_request
  end

  def create(draft_program)
    @draft_program = @work.draft_programs.new(draft_program)
    @draft_program.edit_request.user = current_user

    if draft_program[:program_id].present?
      @program = @work.programs.find(draft_program[:program_id])
      @draft_program.origin = @program
    end

    if @draft_program.valid?
      @draft_program.started_at = @draft_program.started_at - 9.hours
      @draft_program.save(validate: false)
      flash[:notice] = "番組情報の編集リクエストを作成しました"
      redirect_to db_edit_request_path(@draft_program.edit_request)
    else
      render :new
    end
  end

  def edit(id)
    @draft_program = @work.draft_programs.find(id)
  end

  def update(id, draft_program)
    @draft_program = @work.draft_programs.find(id)
    @draft_program.attributes = draft_program

    if @draft_program.valid?
      @draft_program.started_at = @draft_program.started_at - 9.hours
      @draft_program.save(validate: false)
      flash[:notice] = "エピソードの編集リクエストを更新しました"
      redirect_to db_edit_request_path(@draft_program.edit_request)
    else
      render :edit
    end
  end

  private

  def set_work
    @work = Work.find(params[:work_id])
  end
end
