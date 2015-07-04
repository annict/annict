class Db::ItemsController < Db::ApplicationController
  permits :name, :url, :tombo_image, :main

  before_action :load_work, only: [:index, :new, :create, :edit, :update, :destroy]
  before_action :load_item, only: [:edit, :update, :destroy]

  def index
    @items = @work.items
  end

  def new
    @item = @work.items.new
    authorize @item, :new?
  end

  def create(item)
    @item = @work.items.new(item)
    authorize @item, :create?

    if @item.save
      redirect_to db_work_items_path(@work), notice: "作品画像を登録しました"
    else
      render :new
    end
  end

  def edit
    authorize @item, :edit?
  end

  def update(item)
    authorize @item, :update?

    if @item.update_attributes(item)
      redirect_to db_work_items_path(@work), notice: "作品画像を更新しました"
    else
      render :edit
    end
  end

  def destroy(id)
    authorize @item, :destroy?

    @item.destroy

    redirect_to db_work_items_path(@work), notice: "作品画像を削除しました"
  end

  private

  def load_work
    @work = Work.find(params[:work_id])
  end

  def load_item
    @item = @work.items.find(params[:id])
  end
end
