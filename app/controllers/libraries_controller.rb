# frozen_string_literal: true

class LibrariesController < ApplicationV6Controller
  def show
    set_page_category PageCategory::LIBRARY

    @user = User.only_kept.find_by!(username: params[:username])
    @works = @user.works_on(params[:status_kind]).only_kept
    @work_ids = @works.pluck(:id)
    @library_entries = @user.library_entries.where(work_id: @work_ids)
    season_slugs = @works.map(&:season).select(&:present?).map(&:slug).uniq
    @seasons = season_slugs
      .map { |slug| Season.find_by_slug(slug) }
      .sort_by { |s| "#{s.year}#{s.name_value}".to_i }
      .reverse
    @seasons << Season.no_season if @works.with_no_season.present?
    @seasons = Kaminari.paginate_array(@seasons).page(params[:page]).per(8)
  end
end
