# frozen_string_literal: true

class RecordsController < ApplicationController
  def show(username, id)
    @user = User.find_by(username: username)
    @record = @user.records.find(id)
    @work = @record.work
    @episode = @record.episode
    @comments = @record.comments.order(created_at: :desc)
    @comment = Comment.new
  end
end
