# frozen_string_literal: true

class CollectionsController < ApplicationController
  permits :title, :description

  before_action :authenticate_user!, only: %i(edit update destroy)

  def index(page: nil)
    @popular_collections = Collection.published.order(impressions_count: :desc).page(page)
    @newest_collections = Collection.published.order(created_at: :desc).page(page)
    @user_collections = if user_signed_in?
      current_user.
        collections.
        includes(:collection_items).
        published.
        order(updated_at: :desc)
    else
      Collection.none
    end
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
end
