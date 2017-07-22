# frozen_string_literal: true

class UserCollectionsController < ApplicationController
  before_action :load_user, only: %i(index show)

  def index(page: nil)
    @collections = @user.collections.published.order(updated_at: :desc).page(page)
  end

  def show(id)
    @collection = @user.collections.published.find(id)

    impressionist @collection

    @collections = @user.
      collections.
      includes(:collection_items).
      published.
      where.not(id: @collection.id).
      order(updated_at: :desc)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_list",
      user: current_user,
      works: @collection.works
  end

  private

  def load_user
    @user = User.find_by!(username: params[:username])
  end
end
