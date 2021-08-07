# frozen_string_literal: true

class LibrariesController < ApplicationV6Controller
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
    @seasons = Kaminari.paginate_array(@seasons).page(params[:page]).per(8)
  end
end
