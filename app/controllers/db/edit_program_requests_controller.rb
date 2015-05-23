class Db::EditProgramRequestsController < Db::ApplicationController
  before_action :set_work, only: [:new, :create, :edit, :update]
  before_action :set_program, only: [:new, :create, :edit, :update]
  before_action :set_edit_request, only: [:edit, :update]

  def new
    @form = EditRequest::ProgramForm.new
    @form.work = @work
    @form.program = @form.new_attributes = @program
  end

  def create(edit_request_program_form)
    @form = EditRequest::ProgramForm.new(edit_request_program_form)
    @form.user = current_user
    @form.work = @work
    @form.program = @program

    if @form.save
      flash[:notice] = "編集リクエストを送信しました"
      redirect_to db_edit_request_path(@form.edit_request_id)
    else
      render :new
    end
  end

  def edit
    @form = EditRequest::ProgramForm.new
    @form.work = @work
    @form.program = @program
    @form.edit_attributes = @edit_request
  end

  def update(edit_request_program_form)
    @form = EditRequest::ProgramForm.new(edit_request_program_form)
    @form.edit_request_id = @edit_request.id
    @form.user = current_user
    @form.work = @work
    @form.program = @program

    if @form.save
      flash[:notice] = "編集リクエストを更新しました"
      redirect_to db_edit_request_path(@form.edit_request_id)
    else
      render :edit
    end
  end

  private

  def set_work
    @work = Work.find(params[:work_id])
  end

  def set_program
    @program = @work.programs.where(id: params[:program_id]).first
  end

  def set_edit_request
    @edit_request = EditRequest.find(params[:id])
  end
end
