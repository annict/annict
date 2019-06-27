# frozen_string_literal: true

module V3
  class ApplicationController < ActionController::Base
    include HeadersForFastly

    layout "v3"

    before_action :set_vary_header
  end
end
