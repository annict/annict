# frozen_string_literal: true

class LibrariesController < ApplicationController
  before_action :set_user, only: %i(show)
  before_action :set_display_option, only: %i(show)

  def show
    @works = @user.works.on(params[:status_kind]).without_deleted
    season_slugs = @works.map(&:season).select(&:present?).map(&:slug).uniq
    @seasons = season_slugs.
      map { |slug| Season.find_by_slug(slug) }.
      sort_by { |s| "#{s.year}#{s.name_value}".to_i }.
      reverse
    @seasons << Season.no_season if @works.with_no_season.present?
    paginate_per = @display_option == "grid_detailed" ? 8 : 20
    @seasons = Kaminari.paginate_array(@seasons).page(params[:page]).per(paginate_per)

    if @display_option == "grid_detailed"
      @work_tags_data = Work.work_tags_data(@works, @user)
      @work_comment_data = Work.work_comment_data(@works, @user)
    end

    return unless user_signed_in?

    if @display_option.in?(Setting.display_option_user_work_list.values)
      current_user.setting.update_column(:display_option_user_work_list, @display_option)
    end

    store_page_params(works: @seasons.flat_map { |s| s.no_season? ? @works.with_no_season : @works.by_season(s.slug) })
  end

  private

  def set_user
    @user = User.without_deleted.find_by!(username: params[:username])
  end

  def set_display_option
    display_options = Setting.display_option_user_work_list.values
    display = params[:display].in?(display_options) ? params[:display] : nil

    @display_option = if user_signed_in?
      display.presence || current_user.setting.display_option_user_work_list
    else
      display.presence || "grid_detailed"
    end
  end
end
