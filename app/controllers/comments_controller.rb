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
  before_action :set_checkin, only: [:create]


  def create(comment)
    @comment = @checkin.comments.new(comment)
    @comment.user = current_user

    if @comment.save
      redirect_to work_episode_checkin_path(@work, @episode, @checkin), notice: t('comments.saved')
    else
      @comments = @checkin.comments.order(created_at: :desc)
      render template: 'checkins/show'
    end
  end
end
