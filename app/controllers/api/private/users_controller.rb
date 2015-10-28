class Api::Private::UsersController < Api::Private::ApplicationController
  before_action :set_user, only: [:activities]

  def activities(page: nil)
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
