# typed: false
# frozen_string_literal: true

class HomeController < ApplicationV6Controller
  before_action :authenticate_user!

  def show
    set_page_category PageCategory::HOME

    @activity_groups = ActivityGroup
      .preload(user: :profile)
      .where(user_id: current_user.following_user_ids)
      .order(created_at: :desc)
      .page(params[:page])
      .per(30)
      .without_count

    @work_ids = if @activity_groups.present?
      @activity_groups.flat_map.with_prelude { |ags| ags.first_item.work_id }.uniq
    else
      []
    end
  end
end
