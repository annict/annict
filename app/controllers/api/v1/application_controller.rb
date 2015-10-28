class Api::V1::ApplicationController < ActionController::Base
  before_action -> { doorkeeper_authorize! :read }
end
