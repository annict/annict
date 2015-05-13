class Db::EditWorkRequestsController < Db::ApplicationController
  permits :season_id, :sc_tid, :title, :media, :official_site_url, :wikipedia_url,
          :released_at, :released_at_about, :twitter_username, :twitter_hashtag,
          :fetch_syobocal, :edit_request_title, :edit_request_body

  before_action :set_work, only: [:new, :create]
  before_action :set_edit_request, only: [:edit, :update]

  def new
    @form = EditRequest::DraftWorksForm.new

    if @work.present?
      @form.work = @work
      @form.edit_request_resource = @work
    end
  end

  def create(edit_request_draft_works_form)
    @form = EditRequest::DraftWorksForm.new(edit_request_draft_works_form)
    @form.edit_request_resource = @work
    @form.edit_request_user_id = current_user.id

    if @form.save
      flash[:notice] = "編集リクエストを送信しました"
      redirect_to db_edit_request_path(@form._edit_request_id)
    else
      render :new
    end
  end

  def edit
    @form = EditRequest::DraftWorksForm.new
    @form.edit_request = @edit_request
  end

  def update(edit_request_draft_works_form)
    @form = EditRequest::DraftWorksForm.new(edit_request_draft_works_form)
    @form.edit_request_id = @edit_request.id
    @form.edit_request_user_id = @edit_request.user.id
    @form.edit_request_resource = @edit_request.resource

    if @form.save
      flash[:notice] = "編集リクエストを更新しました"
      redirect_to db_edit_request_path(@form._edit_request_id)
    else
      render :edit
    end
  end

  private

  def set_work
    return unless params.has_key?(:work_id)

    @work = Work.find(params[:work_id])
  end

  def set_edit_request
    @edit_request = EditRequest.find(params[:id])
  end
end
