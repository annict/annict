class Api::UsersController < Api::ApplicationController
  before_filter :set_user, only: [:activities]

  def activities(page)
    @activities = @user.activities
                    .includes(:recipient, :trackable, :user)
                    .order(created_at: :desc)
                    .page(page)
  end

  private

  def set_user
    @user = User.find_by!(username: params[:user_id])
  end
end