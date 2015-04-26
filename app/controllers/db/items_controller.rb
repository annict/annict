class Db::ItemsController < Db::ApplicationController
  permits :name, :tombo_image, :main, :url

  before_action :set_work, only: [:index, :new, :create, :edit, :update, :destroy]
  before_action :set_item, only: [:edit, :update, :destroy]


  def index
    @items = @work.items
  end

  def new
    @item = @work.items.new
  end

  def create(item)
    @item = @work.items.new(item)

    if @item.save
      redirect_to db_work_items_path(@work), notice: "作品画像を登録しました"
    else
      render :new
    end
  end

  def update(item)
    if @item.update_attributes(item)
      redirect_to db_work_items_path(@work), notice: "作品画像を更新しました"
    else
      render :edit
    end
  end

  def destroy
    @item.destroy
    redirect_to db_work_items_path(@work), notice: "作品画像を削除しました"
  end

  private

  def set_item
    @item = @work.items.find(params[:id])
  end
end
