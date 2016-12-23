# frozen_string_literal: true
# == Schema Information
#
# Table name: comments
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  checkin_id  :integer          not null
#  body        :text             not null
#  likes_count :integer          default(0), not null
#  created_at  :datetime
#  updated_at  :datetime
#  work_id     :integer
#
# Indexes
#
#  comments_checkin_id_idx    (checkin_id)
#  comments_user_id_idx       (user_id)
#  index_comments_on_work_id  (work_id)
#

class CommentsController < ApplicationController
  permits :body

  before_action :authenticate_user!
  before_action :load_user, only: %i(create)
  before_action :load_record, only: %i(create)

  def create(comment)
    @user = @record.user
    @comment = @record.comments.new(comment)
    @comment.user = current_user

    if @comment.save
      redirect_to record_path(@user.username, @record),
        notice: t("messages.comments.saved")
    else
      @work = @record.work
      @episode = @record.episode
      @comments = @record.comments.order(created_at: :desc)
      render "/records/show"
    end
  end

  private

  def load_user
    @user = User.find_by(username: params[:username])
  end

  def load_record
    @record = @user.records.find(params[:record_id])
  end
end
