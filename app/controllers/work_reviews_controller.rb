# frozen_string_literal: true

class WorkReviewsController < ApplicationController
  before_action :load_work, only: %i(index)

  def index(page: nil)
    @reviews = @work.
      reviews.
      published.
      with_body.
      includes(user: :profile)
    @reviews = localable_resources(@reviews)
    @reviews = @reviews.order(created_at: :desc).page(params[:page])

    @is_spoiler = @user.present? && work_records.present? && @user.hide_work_record_body?(work_records.first.work)

    return unless user_signed_in?

    @review = @work.reviews.new

    store_page_params(work: @work)
  end
end
