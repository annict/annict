# frozen_string_literal: true

class ProfilesController < ApplicationV6Controller
  def show
    set_page_category PageCategory::PROFILE

    @user = User.only_kept.find_by!(username: params[:username])
    @profile = @user.profile

    @activity_groups = @user
      .activity_groups
      .order(created_at: :desc)
      .page(params[:page])
      .per(30)
      .without_count

    @work_ids = if @activity_groups.present?
      @activity_groups.flat_map.with_prelude do |activity_group|
        activity_group.activity_items.first.work_id
      end.uniq
    else
      []
    end
  end
end
