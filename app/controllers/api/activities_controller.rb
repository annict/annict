class Api::ActivitiesController < Api::ApplicationController
  before_action :authenticate_user!, only: [:index]

  def index(page)
    @activities = current_user.following_activities
                    .order(created_at: :desc)
                    .includes(:recipient, :trackable, user: :profile)
                    .page(page)
  end
end