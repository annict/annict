# frozen_string_literal: true
# == Schema Information
#
# Table name: comments
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  record_id  :integer          not null
#  body        :text             not null
#  likes_count :integer          default(0), not null
#  created_at  :datetime
#  updated_at  :datetime
#  work_id     :integer
#
# Indexes
#
#  comments_record_id_idx    (record_id)
#  comments_user_id_idx       (user_id)
#  index_comments_on_work_id  (work_id)
#

class CommentsController < ApplicationController
  before_action :authenticate_user!
  before_action :load_user, only: %i(create)
  before_action :load_record, only: %i(create)
  before_action :load_comment, only: %i(edit update destroy)

  def create
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
    authorize @comment, :edit?
  end

  def update
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
    authorize @comment, :destroy?

    @comment.destroy

    path = record_path(@comment.episode_record.user.username, @comment.episode_record.record)
    redirect_to path, notice: t("messages.comments.deleted")
  end

  private

  def load_user
    @user = User.published.find_by(username: params[:username])
  end

  def load_record
    @record = @user.records.published.find(params[:record_id])
  end

  def load_comment
    @comment = current_user.record_comments.find(params[:id])
  end

  def comment_params
    params.require(:comment).permit(:body)
  end
end
