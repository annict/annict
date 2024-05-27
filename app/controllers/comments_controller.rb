# typed: false
# frozen_string_literal: true

# 現状使われていない
# TODO: https://github.com/kiraka/annict/issues/4 で復活させる
class CommentsController < ApplicationV6Controller
  before_action :authenticate_user!

  def create
    @user = User.only_kept.find_by(username: params[:username])
    @record = @user.records.only_kept.find(params[:record_id])
    @user = @record.user
    @comment = @record.episode_record.comments.new(comment_params)
    @comment.user = current_user
    @comment.work = @record.work
    @comment.detect_locale!(:body)

    if @comment.save
      redirect_to record_path(@user.username, @record),
        notice: t("messages.comments.saved")
    else
      @work = @record.work
      @episode = @record.episode
      @comments = @record.episode_record.comments.order(created_at: :desc)
      render "/records/show"
    end
  end

  def edit
    @comment = current_user.record_comments.find(params[:id])
    authorize @comment, :edit?
  end

  def update
    @comment = current_user.record_comments.find(params[:id])
    authorize @comment, :update?

    @comment.attributes = comment_params
    @comment.detect_locale!(:body)

    if @comment.save
      path = record_path(@comment.episode_record.user.username, @comment.episode_record.record)
      redirect_to path, notice: t("messages.comments.updated")
    else
      render :edit
    end
  end

  def destroy
    @comment = current_user.record_comments.find(params[:id])
    authorize @comment, :destroy?

    @comment.destroy

    path = record_path(@comment.episode_record.user.username, @comment.episode_record.record)
    redirect_to path, notice: t("messages.comments.deleted")
  end

  private

  def comment_params
    params.require(:comment).permit(:body)
  end
end
