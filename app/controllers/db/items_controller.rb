class Db::ItemsController < Db::ApplicationController
  before_action :set_work, only: [:index, :destroy]
  before_action :set_item, only: [:destroy]

  def index
    @items = @work.items
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
