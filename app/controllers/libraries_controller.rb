# frozen_string_literal: true

class LibrariesController < ApplicationV6Controller
  before_action :set_display_option, only: %i[show]

  def show
    set_page_category PageCategory::LIBRARY

    @user = User.only_kept.find_by!(username: params[:username])
    @animes = @user.animes_on(params[:status_kind]).only_kept
    @anime_ids = @animes.pluck(:id)
    season_slugs = @animes.map(&:season).select(&:present?).map(&:slug).uniq
    @seasons = season_slugs
      .map { |slug| Season.find_by_slug(slug) }
      .sort_by { |s| "#{s.year}#{s.name_value}".to_i }
      .reverse
    @seasons << Season.no_season if @animes.with_no_season.present?
    paginate_per = @display_option == "grid_detailed" ? 8 : 20
    @seasons = Kaminari.paginate_array(@seasons).page(params[:page]).per(paginate_per)

    return unless user_signed_in?

    if @display_option.in?(Setting.display_option_user_work_list.values)
      current_user.setting.update_column(:display_option_user_work_list, @display_option)
    end
  end

  private

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
