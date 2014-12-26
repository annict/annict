class Api::UsersController < Api::ApplicationController
  before_filter :authenticate_user!, only: [:share]
  before_filter :set_user, only: [:activities]

  def activities(page: nil)
    @activities = @user.activities
                    .includes(:recipient, :trackable, :user)
                    .order(created_at: :desc)
                    .page(page)
  end

  def share(body)
    TwitterWatchingShareWorker.perform_async(current_user.id, body)

    render status: 200, nothing: true
  end

  private

  def set_user
    @user = User.find_by!(username: params[:user_id])
  end
end
