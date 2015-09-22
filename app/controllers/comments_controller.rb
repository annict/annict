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