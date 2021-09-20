# frozen_string_literal: true

class CollectionsController < ApplicationV6Controller
  def index
    set_page_category PageCategory::COLLECTION_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile
    @collections = @user.collections.only_kept.order(created_at: :desc)
  end
end
