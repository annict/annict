# frozen_string_literal: true

class WorkItemsController < ApplicationController
  before_action :authenticate_user!, only: %i(new destroy)
  before_action :load_work, only: %i(index new destroy)
  before_action :load_i18n, only: %i(new)

  def index
    @items = @work.
      items.
      published
    @items = localable_resources(@items)
    @items = @items.order(created_at: :desc).page(params[:page])

    store_page_params(work: @work)
  end

  def new
    @item = @work.items.new

    store_page_params(work: @work)
  end

  def destroy
    item = @work.items.published.find(params[:id])
    work_item = @work.resource_items.find_by(item: item, user: current_user)

    work_item.destroy

    flash[:notice] = t("messages._common.deleted")
    redirect_back fallback_location: work_items_path(@work)
  end

  private

  def load_i18n
    keys = {
      "messages._components.amazon_item_attacher.error": nil
    }

    load_i18n_into_gon keys
  end
end
