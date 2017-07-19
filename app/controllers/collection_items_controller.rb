# frozen_string_literal: true

class CollectionItemsController < ApplicationController
  permits :title, :comment, :position

  before_action :authenticate_user!, only: %i(edit update destroy)
  before_action :load_collection_item, only: %i(edit update destroy)

  def update(collection_item)
    @collection_item.attributes = collection_item

    return render(:edit) unless @collection_item.save

    flash[:notice] = t("messages._common.updated")
    redirect_to user_collection_path(current_user.username, @collection_item.collection)
  end

  def destroy
    @collection_item.destroy

    flash[:notice] = t("messages._common.deleted")
    redirect_to user_collection_path(current_user.username, @collection_item.collection)
  end

  private

  def load_collection_item
    @collection_item = current_user.
      collection_items.
      published.
      where(collection_id: params[:collection_id]).
      find(params[:id])
  end
end
