# frozen_string_literal: true

module V3
  class ApplicationController < ActionController::Base
    include Localable

    layout "v3"

    helper_method :local_url
  end
end
