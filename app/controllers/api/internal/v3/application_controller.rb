# frozen_string_literal: true

module Api
  module Internal
    module V3
      class ApplicationController < ActionController::Base
        skip_before_action :verify_authenticity_token
      end
    end
  end
end
