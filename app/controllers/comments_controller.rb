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
  before_action :set_work,    only: [:create]
  before_action :set_episode, only: [:create]
  before_action :load_record, only: [:create]

  def create(comment)
    @comment = @record.comments.new(comment)
    @comment.user = current_user

    if @comment.save
      path = work_episode_checkin_path(@work, @episode, @record)
      redirect_to path, notice: t("comments.saved")
    else
      @comments = @record.comments.order(created_at: :desc)
      render template: "checkins/show"
    end
  end
end
