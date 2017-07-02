# frozen_string_literal: true

class WorkItemsController < ApplicationController
  before_action :authenticate_user!, only: %i(new destroy)
  before_action :load_work, only: %i(index new destroy)
  before_action :set_page_object, only: %i(index new)

  def index(page: nil)
    @items = @work.
      items.
      published.
      order(created_at: :desc).
      page(page)
  end

  def new
    @item = @work.items.new
  end

  def destroy(id)
    item = @work.items.published.find(id)
    work_item = @work.resource_items.find_by(item: item, user: current_user)

    work_item.destroy

    flash[:notice] = t("messages._common.deleted")
    redirect_back fallback_location: work_items_path(@work)
  end

  private

  def set_page_object
    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end
end
