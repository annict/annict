# frozen_string_literal: true

module Api
  module Internal
    module V3
      class ApplicationController < ActionController::Base
        include ViewerIdentifiable
      end
    end
  end
end
