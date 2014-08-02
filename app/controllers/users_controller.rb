class UsersController < ApplicationController
  before_filter :set_user, only: [:show, :works]


  def works(status_kind, page)
    @works = @user.works_on(status_kind).page(page)
  end


  private

  def set_user
    @user = User.find_by!(username: params[:id])
  end
end