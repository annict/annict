class Api::V1::UsersController < Api::V1::ApplicationController
  def me
    @user = current_resource_owner
  end

  private

  def current_resource_owner
    User.find(doorkeeper_token.resource_owner_id) if doorkeeper_token
  end
end
