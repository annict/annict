# frozen_string_literal: true

class CollectionsController < ApplicationController
  before_action :authenticate_user!, only: %i(new edit update destroy)

  def index(page: nil)
    @collections = Collection.published.page(page)
  end

  def new
    collection = current_user.collections.create(title: "")
    redirect_to edit_collection_path(collection)
  end

  def edit(id)
    @collection = current_user.collections.find(id)
  end

  def show(id)
    @collection = Collection.published.find(id)
    impressionist @collection
  end
end
