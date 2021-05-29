# frozen_string_literal: true

module V6
  class HomeController < V6::ApplicationController
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

      @anime_ids = if @activity_groups.present?
        @activity_groups.flat_map.with_prelude { |ags| ags.first_item.anime_id }.uniq
      else
        []
      end
    end
  end
end
