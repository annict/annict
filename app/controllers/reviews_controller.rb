# frozen_string_literal: true

class WorkRecordsController < ApplicationController

  before_action :load_user, only: %i(index)
  before_action :load_work, only: %i(new create edit update destroy)

  def index(page: nil)
    @reviews = @user.
      reviews.
      includes(work: :work_image).
      published
    @reviews = localable_resources(@reviews)
    @reviews = @reviews.order(created_at: :desc).page(page)

    return unless user_signed_in?

    @works = Work.where(id: @reviews.pluck(:work_id))

    store_page_params(works: @works)
  end

  def destroy(id)
    @review = current_user.reviews.published.find(id)
    authorize @review, :destroy?

    @review.destroy

    flash[:notice] = t("messages._common.deleted")
    redirect_to work_reviews_path(@review.work)
  end

  private

  def load_user
    @user = User.find_by!(username: params[:username])
  end

end
