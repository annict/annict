class Db::ItemsController < Db::ApplicationController
  before_action :load_work, only: [:index, :destroy]

  def index
    @items = @work.items
  end

  def destroy(id)
    @item = @work.items.find(id)
    authorize @item, :destroy?

    @item.destroy

    redirect_to db_work_items_path(@work), notice: "作品画像を削除しました"
  end

  private

  def load_work
    @work = Work.find(params[:work_id])
  end
end
