# frozen_string_literal: true

class LibrariesController < ApplicationController
  before_action :set_user, only: %i(show)

  def show(status_kind, page: 1)
    @works = @user.works.on(status_kind).published
    season_slugs = @works.map(&:season).select(&:present?).map(&:slug).uniq
    @seasons = season_slugs.
      map { |slug| Season.find_by_slug(slug) }.
      sort_by { |s| "#{s.year}#{s.name_value}".to_i }.
      reverse
    @seasons = Kaminari.paginate_array(@seasons).page(page).per(5)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_list",
      user: current_user,
      works: @seasons.flat_map { |s| @works.by_season(s.slug) }
  end

  private

  def set_user
    @user = User.find_by!(username: params[:username])
  end
end
