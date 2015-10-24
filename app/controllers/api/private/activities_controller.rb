class Api::ActivitiesController < Api::ApplicationController
  before_action :authenticate_user!, only: [:index]

  def index(page: nil)
    @activities = current_user.following_activities
                    .order(id: :desc)
                    .includes(:recipient, trackable: :user, user: :profile)
                    .page(page)
  end
end
