class Db::EditItemRequestsController < Db::ApplicationController
  before_action :set_work, only: [:new, :create, :edit, :update]
  before_action :set_item, only: [:new, :create]
  before_action :set_edit_request, only: [:edit, :update]

  def new
    @form = EditRequest::ItemForm.new
    @form.work = @work
    @form.item = @form.new_attributes = @item if @item.present?
  end

  def create(edit_request_item_form)
    @form = EditRequest::ItemForm.new(edit_request_item_form)
    @form.user = current_user
    @form.work = @work
    @form.item = @item if @item.present?

    if @form.save
      flash[:notice] = "編集リクエストを送信しました"
      redirect_to db_edit_request_path(@form.edit_request_id)
    else
      render :new
    end
  end

  def edit
    @form = EditRequest::ItemForm.new
    @form.work = @work
    @form.item = @edit_request.resource
    @form.edit_attributes = @edit_request
  end

  def update(edit_request_item_form)
    @form = EditRequest::ItemForm.new(edit_request_item_form)
    @form.edit_request_id = @edit_request.id
    @form.user = current_user
    @form.work = @work
    @form.item = @edit_request.resource

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

  def set_item
    @item = @work.items.where(id: params[:item_id]).first
  end

  def set_edit_request
    @edit_request = EditRequest.find(params[:id])
  end
end
