# frozen_string_literal: true

class Oauth::ApplicationsController < Doorkeeper::ApplicationsController
  include ViewSelector

  layout "v3/application"

  before_action :set_search_params

  private

  def set_search_params
    @search = SearchService.new(params[:q])
  end
end
