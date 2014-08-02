class CommentsController < ApplicationController
  permits :body

  before_filter :authenticate_user!
  before_filter :set_work,    only: [:create]
  before_filter :set_episode, only: [:create]
  before_filter :set_checkin, only: [:create]


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