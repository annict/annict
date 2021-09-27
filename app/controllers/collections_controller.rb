# frozen_string_literal: true

class CollectionsController < ApplicationV6Controller
  def index
    set_page_category PageCategory::COLLECTION_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @collections = @user.collections.only_kept.order(created_at: :desc)
  end

  def show
    set_page_category PageCategory::COLLECTION

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @collection = @user.collections.only_kept.find(params[:collection_id])
    @collection_items = @collection.collection_items.only_kept.preload(:work).order(:position)
    @work_ids = @collection_items.pluck(:work_id)
    @library_entry_by_work_id = @user.library_entries.where(work_id: @work_ids).each_with_object({}) { |le, h| h[le.work_id] = le }
  end
end
