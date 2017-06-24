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
      published.
      order(created_at: :desc).
      page(page)
  end

  def show
    @work = @review.work
    @is_spoiler = user_signed_in? && current_user.hide_review?(@review)
    set_page_object
  end

  def new
    @review = @work.reviews.new
    set_page_object
  end

  def create(review)
    @review = @work.reviews.new(review)
    @review.user = current_user

    begin
      @review.save!
      CreateReviewActivityJob.perform_later(current_user.id, @review.id)
      ga_client.page_category = params[:page_category]
      ga_client.events.create(:reviews, :create)
      flash[:notice] = t("messages._common.post")
      redirect_to review_path(current_user.username, @review)
    rescue
      set_page_object
      render :new
    end
  end

  def edit(id)
    @review = @work.reviews.published.find(id)
    authorize @review, :edit?
    set_page_object
  end

  def update(id, review)
    @review = @work.reviews.published.find(id)
    authorize @review, :update?

    @review.modified_at = Time.now

    if @review.update_attributes(review)
      flash[:notice] = t("messages._common.updated")
      redirect_to review_path(@review.user.username, @review)
    else
      set_page_object
      render :edit
    end
  end

  def destroy(id)
    @review = @work.reviews.published.find(id)
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

  def set_page_object
    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end
end
