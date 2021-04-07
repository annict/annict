# frozen_string_literal: true

module My
  class ApplicationController < ActionController::Base
    include PageCategorizable

    layout "simple"

    helper_method :page_category

    before_action :authenticate_user!
  end
end
