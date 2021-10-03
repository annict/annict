# frozen_string_literal: true

class CollectionsController < ApplicationV6Controller
  before_action :authenticate_user!, only: %i[edit update destroy]

  def index
    set_page_category PageCategory::COLLECTION_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @collections = @user.collections.only_kept.order(created_at: :desc)
  end

  def new
    set_page_category PageCategory::NEW_COLLECTION

    @profile = current_user.profile
    @form = Forms::CollectionForm.new
  end

  def create
    @form = Forms::CollectionForm.new(collection_form_params)

    if @form.invalid?
      @profile = current_user.profile

      return render :edit, status: :unprocessable_entity
    end

    result = Creators::CollectionCreator.new(user: current_user, form: @form).call

    flash[:notice] = t "messages._common.created"
    redirect_to user_collection_path(current_user.username, result.collection.id)
  end

  def show
    set_page_category PageCategory::COLLECTION

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @collection = @user.collections.only_kept.find(params[:collection_id])
    @collection_items = @collection.collection_items.only_kept.preload(:user, work: :work_image).order(:position)
    @work_ids = @collection_items.pluck(:work_id)
  end

  def edit
    set_page_category PageCategory::EDIT_COLLECTION

    @profile = current_user.profile
    @collection = current_user.collections.only_kept.find(params[:collection_id])
    @collection_form = Forms::CollectionForm.new(
      collection: @collection,
      name: @collection.name,
      description: @collection.description
    )
  end

  def update
    @collection = current_user.collections.only_kept.find(params[:collection_id])
    @collection_form = Forms::CollectionForm.new(collection_form_params)
    @collection_form.collection = @collection

    if @collection_form.invalid?
      return render :edit, status: :unprocessable_entity
    end

    Updaters::CollectionUpdater.new(user: current_user, form: @collection_form).call

    flash[:notice] = t "messages._common.updated"
    redirect_to user_collection_path(current_user.username, @collection.id)
  end

  def destroy
    collection = current_user.collections.only_kept.find(params[:collection_id])

    Destroyers::CollectionDestroyer.new(collection: collection).call

    flash[:notice] = t "messages._common.deleted"
    redirect_to user_collection_list_path(current_user.username)
  end

  private

  def collection_form_params
    params.required(:forms_collection_form).permit(:name, :description)
  end
end
