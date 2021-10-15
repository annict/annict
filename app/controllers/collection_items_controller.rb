# frozen_string_literal: true

class CollectionItemsController < ApplicationV6Controller
  before_action :authenticate_user!, only: %i[edit update destroy]

  def edit
    set_page_category PageCategory::EDIT_COLLECTION_ITEM

    @profile = current_user.profile
    @collection_item = current_user.collection_items.only_kept.find(params[:collection_item_id])
    @collection = @collection_item.collection
    @work = @collection_item.work
    @form = Forms::CollectionItemForm.new(
      collection_item: @collection_item,
      body: @collection_item.body
    )
  end

  def update
    @collection_item = current_user.collection_items.only_kept.find(params[:collection_item_id])
    @collection = @collection_item.collection
    @work = @collection_item.work
    @form = Forms::CollectionItemForm.new(collection_item_form_params)
    @form.collection_item = @collection_item

    if @form.invalid?
      return render :edit, status: :unprocessable_entity
    end

    Updaters::CollectionItemUpdater.new(user: current_user, form: @form).call

    flash[:notice] = t "messages._common.updated"
    redirect_to user_collection_path(current_user.username, @collection_item.collection_id)
  end

  def destroy
    collection_item = current_user.collection_items.only_kept.find(params[:collection_item_id])

    Destroyers::CollectionItemDestroyer.new(collection_item: collection_item).call

    flash[:notice] = t "messages._common.deleted"
    redirect_to user_collection_path(current_user.username, collection_item.collection_id)
  end

  private

  def collection_item_form_params
    params.required(:forms_collection_item_form).permit(:body)
  end
end
