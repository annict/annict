# frozen_string_literal: true

class CollectionsController < ApplicationController
  impressionist actions: %i(show)

  permits :title, :description

  before_action :authenticate_user!, only: %i(edit update destroy)
  before_action :load_user, only: %i(index show)

  def index(page: nil)
    @collections = @user.collections.published.order(updated_at: :desc).page(page)
  end

  def show(id)
    @collection = @user.collections.published.find(id)
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

  def edit(id)
    @collection = current_user.collections.published.find(id)
  end

  def update(id, collection)
    @collection = current_user.collections.published.find(id)
    @collection.attributes = collection

    return render(:edit) unless @collection.save

    flash[:notice] = t("messages._common.updated")
    redirect_to user_collection_path(current_user.username, @collection)
  end

  def destroy(id)
    @collection = current_user.collections.published.find(id)
    @collection.destroy

    flash[:notice] = t("messages._common.deleted")
    redirect_to user_collections_path(current_user.username)
  end

  private

  def load_user
    @user = User.find_by!(username: params[:username])
  end
end
