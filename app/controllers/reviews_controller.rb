# frozen_string_literal: true

class ReviewsController < ApplicationController
  permits :title, :body, :rating_animation_state, :rating_music_state, :rating_story_state,
    :rating_character_state, :rating_overall_state

  impressionist actions: %i(show)

  before_action :authenticate_user!, only: %i(new create edit update destroy)
  before_action :load_user, only: %i(index show)
  before_action :load_work, only: %i(new create edit update destroy)
  before_action :load_review, only: %i(show)

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

  def show
    @work = @review.work
    @is_spoiler = user_signed_in? && current_user.hide_review?(@review)
    @reviews = @user.
      reviews.
      published.
      where.not(id: @review.id).
      includes(work: :work_image).
      order(id: :desc)
    store_page_params(work: @work)
  end

  def create(review, page: nil)
    @review = @work.reviews.new(review)
    @review.user = current_user
    current_user.setting.attributes = setting_params
    ga_client.page_category = params[:page_category]

    service = NewReviewService.new(current_user, @review, current_user.setting)
    service.ga_client = ga_client
    service.page_category = params[:page_category]
    service.via = "web"

    begin
      service.save!
      flash[:notice] = t("messages._common.post")
      redirect_to review_path(current_user.username, @review)
    rescue
      @reviews = @work.
        reviews.
        published.
        with_body.
        includes(user: :profile)
      @reviews = localable_resources(@reviews)
      @reviews = @reviews.order(created_at: :desc).page(params[:page])

      @is_spoiler = @user.present? && reviews.present? && @user.hide_review?(reviews.first)

      store_page_params(work: @work)

      render "work_reviews/index"
    end
  end

  def edit(id)
    @review = current_user.reviews.published.find(id)
    authorize @review, :edit?
    store_page_params(work: @work)
  end

  def update(id, review)
    @review = current_user.reviews.published.find(id)
    authorize @review, :update?

    @review.attributes = review
    @review.detect_locale!(:body)
    @review.modified_at = Time.now
    current_user.setting.attributes = setting_params

    begin
      ActiveRecord::Base.transaction do
        @review.save!
        current_user.setting.save!
      end
      if current_user.setting.share_review_to_twitter?
        ShareReviewToTwitterJob.perform_later(current_user.id, @review.id)
      end
      if current_user.setting.share_review_to_facebook?
        ShareReviewToFacebookJob.perform_later(current_user.id, @review.id)
      end
      flash[:notice] = t("messages._common.updated")
      redirect_to review_path(@review.user.username, @review)
    rescue
      store_page_params(work: @work)
      render :edit
    end
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

  def load_review
    @review = @user.reviews.published.find(params[:id])
  end

  def setting_params
    params.require(:setting).permit(:share_review_to_twitter, :share_review_to_facebook)
  end
end
