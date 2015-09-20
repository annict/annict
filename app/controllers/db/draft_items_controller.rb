class Db::DraftItemsController < Db::ApplicationController
  permits :name, :url, :tombo_image, :item_id, :main,
          edit_request_attributes: [:id, :title, :body]

  before_action :authenticate_user!
  before_action :set_work, only: [:new, :create, :edit, :update]

  def new(item_id: nil)
    if item_id.present?
      @item = @work.item
      attributes = @item.attributes.slice(*Item::DIFF_FIELDS.map(&:to_s))
      @draft_item = @work.draft_items.new(attributes)
      @draft_item.origin = @item
    else
      @draft_item = @work.draft_items.new
    end

    @draft_item.build_edit_request
  end

  def create(draft_item)
    @draft_item = @work.draft_items.new(draft_item)
    @draft_item.edit_request.user = current_user

    if draft_item[:item_id].present?
      @item = @work.item
      @draft_item.origin = @item
      @draft_item.tombo_image = @item.tombo_image if draft_item[:tombo_image].blank?
    end

    if @draft_item.save
      flash[:notice] = "作品画像の編集リクエストを作成しました"
      redirect_to db_edit_request_path(@draft_item.edit_request)
    else
      render :new
    end
  end

  def edit(id)
    @draft_item = @work.draft_items.find(id)
    authorize @draft_item, :edit?
  end

  def update(id, draft_item)
    @draft_item = @work.draft_items.find(id)
    authorize @draft_item, :update?

    if @draft_item.update(draft_item)
      flash[:notice] = "作品画像の編集リクエストを更新しました"
      redirect_to db_edit_request_path(@draft_item.edit_request)
    else
      render :edit
    end
  end

  private

  def set_work
    @work = Work.find(params[:work_id])
  end
end
