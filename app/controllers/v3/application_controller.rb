# frozen_string_literal: true

module V3
  class ApplicationController < ActionController::Base
    include HeadersForFastly

    layout "v3"

    before_action :set_vary_header

    def render_404
      render file: "#{Rails.root}/public/404", layout: false, status: :not_found
    end
  end
end
