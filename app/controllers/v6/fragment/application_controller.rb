# frozen_string_literal: true

module Fragment
  class ApplicationController < ActionController::Base
    include PageCategorizable

    layout false

    helper_method :page_category
  end
end
