# frozen_string_literal: true

module Api
  module Internal
    class ApplicationController < ActionController::Base
      include ViewerIdentifiable
    end
  end
end
