class Db::ItemsController < Db::ApplicationController
  permits :name, :url, :tombo_image

  before_action :authenticate_user!
  before_action :load_work, only: [:show, :new, :create, :edit, :update, :destroy]
  before_action :load_item, only: [:edit, :update, :destroy]

  def show
    @item = @work.item
  end

  def new
    @item = @work.build_item
    authorize @item, :new?
  end

  def create(item)
    @item = @work.build_item(item)
    authorize @item, :create?

    if @item.save_and_create_db_activity(current_user, "items.create")
      redirect_to db_work_item_path(@work), notice: "作品画像を登録しました"
    else
      render :new
    end
  end

  def edit
    authorize @item, :edit?
  end

  def update(item)
    authorize @item, :update?

    @item.attributes = item
    if @item.save_and_create_db_activity(current_user, "items.update")
      redirect_to db_work_item_path(@work), notice: "作品画像を更新しました"
    else
      render :edit
    end
  end

  def destroy
    authorize @item, :destroy?

    @item.destroy

    redirect_to db_work_item_path(@work), notice: "作品画像を削除しました"
  end

  private

  def load_work
    @work = Work.find(params[:work_id])
  end

  def load_item
    @item = @work.item
    raise ActiveRecord::RecordNotFound if @item.blank?
  end
end
