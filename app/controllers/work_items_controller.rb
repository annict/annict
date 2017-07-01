# frozen_string_literal: true

class WorkItemsController < ApplicationController
  before_action :load_work, only: %i(index new create destroy)
  before_action :set_page_object, only: %i(index new create destroy)

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

  private

  def set_page_object
    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end
end
