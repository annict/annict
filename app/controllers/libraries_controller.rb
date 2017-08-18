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
    @work_tags_data = Work.work_tags_data(@works, @user)
    @work_comment_data = Work.work_comment_data(@works, @user)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_list",
      user: current_user,
      works: @seasons.flat_map { |s| @works.by_season(s.slug) },
      with_friends: false
  end

  private

  def set_user
    @user = User.find_by!(username: params[:username])
  end
end
