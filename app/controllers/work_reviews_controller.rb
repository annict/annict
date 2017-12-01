# frozen_string_literal: true

class WorkReviewsController < ApplicationController
  before_action :load_work, only: %i(index)

  def index(page: nil)
    @reviews = @work.
      reviews.
      published.
      includes(user: :profile).
      order(created_at: :desc).
      page(page)
    @is_spoiler = user_signed_in? && @reviews.present? && current_user.hide_review?(@reviews.first)

    return unless user_signed_in?

    gon.workListData = render_jb "works/_detail",
      user: current_user,
      work: @work
  end
end
