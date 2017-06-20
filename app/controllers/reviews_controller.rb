# frozen_string_literal: true

class ReviewsController < ApplicationController
  permits :title, :body, :rating_animation_state, :rating_music_state, :rating_story_state,
    :rating_character_state, :rating_overall_state

  impressionist actions: %i(show)

  before_action :authenticate_user!, only: %i(create edit update destroy)
  before_action :load_user, only: %i(index create show edit update destroy)
  before_action :load_review, only: %i(show edit update destroy)

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
    @review_comments = @review.review_comments.order(created_at: :desc)
    @review_comment = ReviewComment.new
    @is_spoiler = user_signed_in? && current_user.hide_review?(@review)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end

  def create(checkin)
    @episode = Episode.published.find(checkin[:episode_id])
    @work = @episode.work
    @review = @episode.reviews.new(checkin)
    keen_client.page_category = params[:page_category]
    ga_client.page_category = params[:page_category]

    service = NewRecordService.new(current_user, @review)
    service.keen_client = keen_client
    service.ga_client = ga_client

    begin
      service.save!
      flash[:notice] = t("messages.reviews.created")
      redirect_to work_episode_path(@work, @episode)
    rescue
      service = RecordsListService.new(current_user, @episode, params)

      @all_reviews = service.all_reviews
      @all_comment_reviews = service.all_comment_reviews
      @friend_comment_reviews = service.friend_comment_reviews
      @my_reviews = service.my_reviews
      @selected_comment_reviews = service.selected_comment_reviews

      data = {
        reviewsSortTypes: Setting.reviews_sort_type.options,
        currentRecordsSortType: current_user&.setting&.reviews_sort_type.presence || "created_at_desc",
        pageObject: render_jb("works/_detail", user: current_user, work: @work)
      }
      gon.push(data)

      @is_spoiler = current_user.hide_checkin_comment?(@episode)

      render "/episodes/show"
    end
  end

  def edit
    authorize @review, :edit?
    @work = @review.work
  end

  def update(checkin)
    authorize @review, :update?

    @review.modify_comment = true

    if @review.update_attributes(checkin)
      @review.update_share_checkin_status
      @review.share_to_sns
      path = review_path(@user.username, @review)
      redirect_to path, notice: t("messages.reviews.updated")
    else
      @work = @review.work
      render :edit
    end
  end

  def destroy
    authorize @review, :destroy?

    @review.destroy

    path = work_episode_path(@review.work, @review.episode)
    redirect_to path, notice: t("messages.reviews.deleted")
  end

  def switch(episode_id, to)
    episode = Episode.find(episode_id)
    redirect = redirect_back fallback_location: work_episode_path(episode.work, episode)

    return redirect unless to.in?(Setting.display_option_review_list.values)

    current_user.setting.update_column(:display_option_review_list, to)
    redirect
  end

  private

  def load_user
    @user = User.find_by!(username: params[:username])
  end

  def load_review
    @review = @user.reviews.find(params[:id])
  end
end
